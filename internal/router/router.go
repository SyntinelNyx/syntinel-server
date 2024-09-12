package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	Logger "github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/utils"
)

type Router struct {
	router *chi.Mux
	logger *zap.Logger
}

func SetupRouter() *Router {
	router := chi.NewRouter()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	router.Use(Logger.LoggerMiddleware(logger))

	r := Router{
		router: router,
		logger: logger,
	}

	r.router.Get("/coffee", func(w http.ResponseWriter, r *http.Request) {
		utils.RespondWithError(w, http.StatusTeapot, "I'm a teapot")
	})

	return &r
}

func (r *Router) GetRouter() *chi.Mux {
	return r.router
}
