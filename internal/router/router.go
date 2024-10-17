package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
				utils.RespondWithError(w, http.StatusTeapot, "I'm a teapot")
			})
		})

		apiRouter.Group(func(subRouter chi.Router) {
			subRouter.Use(r.rateLimiter.RateLimitMiddleware(rate.Every(1*time.Second), 3))

			authHandler := auth.NewHandler(r.queries)

			subRouter.Post("/auth/login", authHandler.Login)
			subRouter.Post("/auth/register", authHandler.Register)
		})

		apiRouter.Group(func(subRouter chi.Router) {
			subRouter.Use(r.rateLimiter.RateLimitMiddleware(rate.Every(1*time.Second), 3))

			authHandler := auth.NewHandler(r.queries)
			subRouter.Use(authHandler.JWTMiddleware)

			subRouter.Get("/auth/validate", func(w http.ResponseWriter, req *http.Request) {
				utils.RespondWithJSON(w, http.StatusOK,
					map[string]string{"message": auth.GetClaims(req.Context()).AccountID})
			})
		})
	})

	return &r
}

func (r *Router) GetRouter() *chi.Mux {
	return r.router
}
