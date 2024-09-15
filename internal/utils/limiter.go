package utils

import (
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiters map[string]map[string]*rate.Limiter // map[path][ip] => limiter
	mu       sync.RWMutex
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]map[string]*rate.Limiter),
	}
}

func GetIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func (rl *RateLimiter) GetLimiter(path string, ip string, r rate.Limit, burst int) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if _, exists := rl.limiters[path]; !exists {
		rl.limiters[path] = make(map[string]*rate.Limiter)
	}

	if limiter, exists := rl.limiters[path][ip]; exists {
		return limiter
	}

	limiter := rate.NewLimiter(r, burst)
	rl.limiters[path][ip] = limiter
	return limiter
}

func (rl *RateLimiter) RateLimitMiddleware(rateLimit rate.Limit, burst int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := GetIP(r)
			limiter := rl.GetLimiter(r.URL.Path, ip, rateLimit, burst)

			if !limiter.Allow() {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
