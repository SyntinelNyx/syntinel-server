package limiter

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"

	"github.com/SyntinelNyx/syntinel-server/internal/response"
)

type RateLimiter struct {
	limiters map[string]map[string]*rate.Limiter // map[path][ip] => limiter
	mu       sync.RWMutex
}

func New() *RateLimiter {
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

func (rl *RateLimiter) Get(path string, ip string, r rate.Limit, burst int) *rate.Limiter {
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

func (rl *RateLimiter) Middleware(rateLimit rate.Limit, burst int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := GetIP(r)
			limiter := rl.Get(r.URL.Path, ip, rateLimit, burst)

			if !limiter.Allow() {
				response.RespondWithError(w, r, http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests), fmt.Errorf("too many requests"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
