package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/SyntinelNyx/syntinel-server/internal/asset"
	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/limiter"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/SyntinelNyx/syntinel-server/internal/role"
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

			assetHandler := asset.NewHandler(r.queries)

			subRouter.Get("/coffee", func(w http.ResponseWriter, r *http.Request) {
				response.RespondWithJSON(w, http.StatusTeapot, map[string]string{"error": "I'm A Teapot!"})
			})
			subRouter.Post("/agent/enroll", assetHandler.Enroll)
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
			assetHandler := asset.NewHandler(r.queries)

			subRouter.Use(authHandler.JWTMiddleware)
			subRouter.Use(authHandler.CSRFMiddleware)

			subRouter.Get("/assets", assetHandler.Retrieve)

			subRouter.Post("/role/retrieve", roleHandler.Retrieve)
			subRouter.Post("/role/create", roleHandler.Create)
			subRouter.Post("/role/delete", roleHandler.DeleteRole)
		})

		apiRouter.Group(func(subRouter chi.Router) {
			subRouter.Use(r.rateLimiter.Middleware(rate.Every(1*time.Second), 10))

			authHandler := auth.NewHandler(r.queries)
			subRouter.Use(authHandler.JWTMiddleware)
			subRouter.Use(authHandler.CSRFMiddleware)

			subRouter.Get("/auth/validate", func(w http.ResponseWriter, r *http.Request) {
				account := auth.GetClaims(r.Context())
				val, err := account.AccountID.Value()
				if err != nil {
					response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to parse UUID", err)
					return
				}
				response.RespondWithJSON(w, http.StatusOK,
					map[string]string{"accountId": val.(string), "accountType": account.AccountType, "accountUser": account.AccountUser})
			})
			subRouter.Post("/auth/logout", authHandler.Logout)
		})
	})

	return &r
}

func (r *Router) GetRouter() *chi.Mux {
	return r.router
}
