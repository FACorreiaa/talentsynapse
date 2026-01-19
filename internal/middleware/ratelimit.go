package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"
)

// RateLimiter returns a rate limiting middleware
// requestLimit: number of requests allowed
// windowLength: time window for the limit
func RateLimiter(requestLimit int, windowLength time.Duration) func(next http.Handler) http.Handler {
	return httprate.Limit(
		requestLimit,
		windowLength,
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		}),
	)
}

// RateLimitByIP returns rate limiting middleware that limits by IP address only
func RateLimitByIP(requestLimit int, windowLength time.Duration) func(next http.Handler) http.Handler {
	return httprate.LimitByIP(requestLimit, windowLength)
}

// RateLimitByRealIP returns rate limiting middleware using X-Real-IP or X-Forwarded-For
func RateLimitByRealIP(requestLimit int, windowLength time.Duration) func(next http.Handler) http.Handler {
	return httprate.LimitByRealIP(requestLimit, windowLength)
}

// DefaultRateLimiter returns a rate limiter with sensible defaults (100 req/min per IP)
func DefaultRateLimiter() func(next http.Handler) http.Handler {
	return RateLimitByIP(100, time.Minute)
}

// StrictRateLimiter returns a stricter rate limiter (20 req/min per IP)
func StrictRateLimiter() func(next http.Handler) http.Handler {
	return RateLimitByIP(20, time.Minute)
}

// APIRateLimiter returns a rate limiter suitable for API endpoints (1000 req/min per IP)
func APIRateLimiter() func(next http.Handler) http.Handler {
	return RateLimitByIP(1000, time.Minute)
}
