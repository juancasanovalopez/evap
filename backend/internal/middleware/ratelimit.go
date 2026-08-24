package middleware

import (
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimit applies a basic per-IP token bucket limiter. It is "best effort"
// in Lambda: state is only shared across invocations handled by the same
// warm execution environment, not globally. For hard limits, pair this with
// API Gateway stage throttling.
func RateLimit(requestsPerSecond float64, burst int) func(http.Handler) http.Handler {
	limiters := &sync.Map{}

	getLimiter := func(key string) *rate.Limiter {
		if l, ok := limiters.Load(key); ok {
			return l.(*rate.Limiter)
		}
		l, _ := limiters.LoadOrStore(key, rate.NewLimiter(rate.Limit(requestsPerSecond), burst))
		return l.(*rate.Limiter)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !getLimiter(clientIP(r)).Allow() {
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
