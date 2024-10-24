package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/utils"
)

type Router struct {
	router      *chi.Mux
	queries     *query.Queries
	logger      *zap.Logger
	rateLimiter *utils.RateLimiter
}

func SetupRouter(q *query.Queries) *Router {
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	rl := utils.NewRateLimiter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	router.Use(utils.LoggerMiddleware(logger))

	r := Router{
		router:      router,
		queries:     q,
		logger:      logger,
		rateLimiter: rl,
	}

	r.router.Route("/v1/api", func(apiRouter chi.Router) {
		apiRouter.Group(func(subRouter chi.Router) {
			subRouter.Use(r.rateLimiter.RateLimitMiddleware(rate.Every(1*time.Second), 30))

			subRouter.Get("/coffee", func(w http.ResponseWriter, req *http.Request) {
				utils.RespondWithJSON(w, http.StatusTeapot, map[string]string{"error": "I'm A Teapot!"})
			})
		})

		apiRouter.Group(func(subRouter chi.Router) {
			subRouter.Use(r.rateLimiter.RateLimitMiddleware(rate.Every(1*time.Second), 3))

			authHandler := auth.NewHandler(r.queries)

			subRouter.Post("/auth/login", authHandler.Login)
			subRouter.Post("/auth/register", authHandler.Register)
		})

		apiRouter.Group(func(subRouter chi.Router) {
			subRouter.Use(r.rateLimiter.RateLimitMiddleware(rate.Every(1*time.Second), 10))

			authHandler := auth.NewHandler(r.queries)
			subRouter.Use(authHandler.JWTMiddleware)
			subRouter.Use(authHandler.CSRFMiddleware)

			subRouter.Get("/auth/validate", func(w http.ResponseWriter, req *http.Request) {
				account := auth.GetClaims(req.Context())
				utils.RespondWithJSON(w, http.StatusOK,
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
