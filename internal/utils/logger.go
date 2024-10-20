package utils

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type errorKeyType struct{}

var errorContextKey = errorKeyType{}

func SetError(r *http.Request, err error) {
	ctx := context.WithValue(r.Context(), errorContextKey, err)
	*r = *r.WithContext(ctx)
}

func GetError(r *http.Request) error {
	if err, ok := r.Context().Value(errorContextKey).(error); ok {
		return err
	}
	return nil
}

func LoggerMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			logFields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("query", r.URL.RawQuery),
				zap.Int("status", ww.Status()),
				zap.String("ip", r.RemoteAddr),
				zap.String("user-agent", r.UserAgent()),
				zap.Duration("latency", time.Since(start)),
				zap.String("time", time.Now().Format(time.RFC3339)),
			}

			if err := GetError(r); err != nil {
				logFields = append(logFields, zap.Error(err))
			}

			if ww.Status() >= http.StatusBadRequest && ww.Status() != http.StatusTeapot {
				logger.Error("Request failed", logFields...)
			} else {
				logger.Info("Request", logFields...)
			}
		})
	}
}
