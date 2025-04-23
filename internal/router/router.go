package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/SyntinelNyx/syntinel-server/internal/agent"
	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/limiter"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/SyntinelNyx/syntinel-server/internal/role"
	"github.com/SyntinelNyx/syntinel-server/internal/snapshots"
)

type Router struct {
	router      *chi.Mux
	queries     *query.Queries
	logger      *zap.Logger
	rateLimiter *limiter.RateLimiter
}

func SetupRouter(q *query.Queries, origins []string) *Router {
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	zlogger, _ := zap.NewProduction()
	defer zlogger.Sync()

	rl := limiter.New()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	router.Use(logger.Middleware(zlogger))

	r := Router{
		router:      router,
		queries:     q,
		logger:      zlogger,
		rateLimiter: rl,
	}

	r.router.Route("/v1/api", func(apiRouter chi.Router) {
		apiRouter.Group(func(subRouter chi.Router) {
			subRouter.Use(r.rateLimiter.Middleware(rate.Every(1*time.Second), 30))

			agentHandler := agent.NewHandler(r.queries)

			subRouter.Get("/coffee", func(w http.ResponseWriter, req *http.Request) {
				response.RespondWithJSON(w, http.StatusTeapot, map[string]string{"error": "I'm A Teapot!"})
			})
			subRouter.Post("/agent/enroll", agentHandler.Enroll)
		})

		apiRouter.Group(func(subRouter chi.Router) {
			subRouter.Use(r.rateLimiter.Middleware(rate.Every(1*time.Second), 3))

			authHandler := auth.NewHandler(r.queries)

			subRouter.Post("/auth/login", authHandler.Login)
			subRouter.Post("/auth/register", authHandler.Register)
		})

		apiRouter.Group(func(subRouter chi.Router) {
			subRouter.Use(r.rateLimiter.Middleware(rate.Every(1*time.Second), 3))

			roleHandler := role.NewHandler(r.queries)
			authHandler := auth.NewHandler(r.queries)
			subRouter.Use(authHandler.JWTMiddleware)
			subRouter.Use(authHandler.CSRFMiddleware)

			subRouter.Get("/auth/validate", func(w http.ResponseWriter, req *http.Request) {
				account := auth.GetClaims(req.Context())
				response.RespondWithJSON(w, http.StatusOK,
					map[string]string{"account_id": account.AccountID, "account_type": account.AccountType})
			})

			subRouter.Post ("/snapshots/create", snapshots.CreateSnapshot)
			subRouter.Post("/snapshots/list", snapshots.RetrieveAllSnapshots)
			subRouter.Post("/snapshots/restore", snapshots.RestoreSnapshot)


			subRouter.Post("/role/retrieve", roleHandler.Retrieve)
			subRouter.Post("/role/create", roleHandler.Create)
			subRouter.Post("/role/delete", roleHandler.DeleteRole)
		})

		apiRouter.Group(func(subRouter chi.Router) {
			subRouter.Use(r.rateLimiter.Middleware(rate.Every(1*time.Second), 10))

			authHandler := auth.NewHandler(r.queries)
			subRouter.Use(authHandler.JWTMiddleware)
			subRouter.Use(authHandler.CSRFMiddleware)

			subRouter.Get("/auth/validate", func(w http.ResponseWriter, req *http.Request) {
				account := auth.GetClaims(req.Context())
				response.RespondWithJSON(w, http.StatusOK,
					map[string]string{"account_id": account.AccountID, "account_type": account.AccountType})
			})
			subRouter.Post("/auth/logout", authHandler.Logout)
		})
	})

	return &r
}

func (r *Router) GetRouter() *chi.Mux {
	return r.router
}
