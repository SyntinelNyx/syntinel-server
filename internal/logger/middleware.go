package logger

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func LoggerMiddleware(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			defer func() {
				duration := time.Since(start)

				logger.Info("Request",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("query", r.URL.RawQuery),
					zap.Int("status", ww.Status()),
					zap.String("ip", r.RemoteAddr),
					zap.String("user-agent", r.UserAgent()),
					zap.Duration("latency", duration),
					zap.String("time", time.Now().Format(time.RFC3339)),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
