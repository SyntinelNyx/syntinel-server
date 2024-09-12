package router

import (
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
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

	r := Router{
		router: router,
		logger: logger,
	}

	return &r
}

func (r *Router) GetRouter() *chi.Mux {
	return r.router
}
