package limiter

import (
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestGetIP(t *testing.T) {
	req1 := &http.Request{RemoteAddr: "192.168.1.100:12345"}
	ip1 := GetIP(req1)
	assert.Equal(t, "192.168.1.100", ip1)

	req2 := &http.Request{RemoteAddr: "invalid-address"}
	ip2 := GetIP(req2)
	assert.Equal(t, "invalid-address", ip2)
}

func TestGetLimiter(t *testing.T) {
	rl := New()

	limiter1 := rl.Get("/test", "127.0.0.1", 1, 1)
	require.NotNil(t, limiter1)

	limiter2 := rl.Get("/test", "127.0.0.1", 1, 1)
	assert.Same(t, limiter1, limiter2)

	limiter3 := rl.Get("/another", "127.0.0.1", 1, 1)
	assert.NotSame(t, limiter1, limiter3)

	limiter4 := rl.Get("/test", "192.168.0.1", 1, 1)
	assert.NotSame(t, limiter1, limiter4)
}

func TestMiddleware(t *testing.T) {
	rl := New()
	rateLimit := rate.Limit(1)
	burst := 1

	var nextCalled int32

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.StoreInt32(&nextCalled, 1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	handler := rl.Middleware(rateLimit, burst)(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = net.JoinHostPort("127.0.0.1", "12345")

	rr := httptest.NewRecorder()
	atomic.StoreInt32(&nextCalled, 0)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, int32(1), atomic.LoadInt32(&nextCalled))

	rr = httptest.NewRecorder()
	atomic.StoreInt32(&nextCalled, 0)
	handler.ServeHTTP(rr, req)

	assert.NotEqual(t, http.StatusOK, rr.Code)

	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	assert.Equal(t, int32(0), atomic.LoadInt32(&nextCalled))
}

func TestMiddlewareIndependent(t *testing.T) {
	rl := New()
	rateLimit := rate.Limit(1)
	burst := 1

	var handlerCallCount int32

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&handlerCallCount, 1)
		w.WriteHeader(http.StatusOK)
	})

	middleware := rl.Middleware(rateLimit, burst)

	req1 := httptest.NewRequest(http.MethodGet, "/path1", nil)
	req1.RemoteAddr = net.JoinHostPort("127.0.0.1", "1111")

	req2 := httptest.NewRequest(http.MethodGet, "/path2", nil)
	req2.RemoteAddr = net.JoinHostPort("127.0.0.2", "2222")

	rr1 := httptest.NewRecorder()
	middleware(nextHandler).ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusOK, rr1.Code)

	rr2 := httptest.NewRecorder()
	middleware(nextHandler).ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)

	assert.Equal(t, int32(2), atomic.LoadInt32(&handlerCallCount))
}

func TestLimiterConcurrency(t *testing.T) {
	rl := New()
	path := "/concurrent"
	ip := "127.0.0.1"

	const goroutines = 50
	limiterCh := make(chan *rate.Limiter, goroutines)
	doneCh := make(chan struct{})

	for i := 0; i < goroutines; i++ {
		go func() {
			limiter := rl.Get(path, ip, 5, 10)
			limiterCh <- limiter
			doneCh <- struct{}{}
		}()
	}

	for i := 0; i < goroutines; i++ {
		<-doneCh
	}

	close(limiterCh)
	var firstLimiter *rate.Limiter
	for lim := range limiterCh {
		if firstLimiter == nil {
			firstLimiter = lim
		} else {
			assert.Equal(t, firstLimiter, lim)
		}
	}
}
