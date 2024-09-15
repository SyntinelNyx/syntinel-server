package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/SyntinelNyx/syntinel-server/internal/utils"
)

type Router struct {
	router      *chi.Mux
	logger      *zap.Logger
	rateLimiter *utils.RateLimiter
}

func SetupRouter() *Router {
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
		logger:      logger,
		rateLimiter: rl,
	}

	r.router.Group(func(r chi.Router) {
		r.Use(rl.RateLimitMiddleware(rate.Every(1*time.Second), 30))
		r.Get("/coffee", func(w http.ResponseWriter, r *http.Request) {
			utils.RespondWithError(w, http.StatusTeapot, "I'm a teapot")
		})
	})

	return &r
}

func (r *Router) GetRouter() *chi.Mux {
	return r.router
}
