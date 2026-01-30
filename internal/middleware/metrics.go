// Package middleware provides HTTP middleware for the application.
package middleware

import (
	"net/http"
	"time"

	"github.com/FACorreiaa/talentsynapse/internal/observability"
)

// MetricsMiddleware tracks HTTP request metrics in Sentry
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Sanitize path for consistent metrics
		sanitizedPath := sanitizePathForMetrics(r.URL.Path)

		// Track request count
		observability.IncrementCounter("http.requests", map[string]string{
			"method": r.Method,
			"path":   sanitizedPath,
		})

		// Wrap ResponseWriter to capture status code
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(ww, r)

		// Track response time in milliseconds
		duration := time.Since(start).Milliseconds()
		observability.TrackMetric("http.response_time", float64(duration), map[string]string{
			"method": r.Method,
			"path":   sanitizedPath,
			"status": http.StatusText(ww.statusCode),
		})

		// Track request size (payload)
		if r.ContentLength > 0 {
			observability.TrackMetric("http.request_size", float64(r.ContentLength), map[string]string{
				"method": r.Method,
				"path":   sanitizedPath,
			})
		}

		// Track status code distribution
		observability.IncrementCounter("http.status", map[string]string{
			"code":   http.StatusText(ww.statusCode),
			"method": r.Method,
			"path":   sanitizedPath,
		})
	})
}

// sanitizePathForMetrics sanitizes URL paths for use in metrics
// Less strict than Prometheus version since Sentry is more forgiving
func sanitizePathForMetrics(path string) string {
	if path == "" {
		return "/"
	}

	// Limit length to prevent excessive data
	const maxPathLength = 200
	if len(path) > maxPathLength {
		path = path[:maxPathLength] + "..."
	}

	return path
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
