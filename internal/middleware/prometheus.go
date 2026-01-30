// Package middleware provides HTTP middleware for the application.
package middleware

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/FACorreiaa/talentsynapse/internal/observability"
)

// PrometheusMiddleware tracks detailed HTTP metrics for Prometheus
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Sanitize path for Prometheus labels
		sanitizedPath := sanitizePathForPrometheus(r.URL.Path)

		// Track in-flight requests
		observability.HTTPInFlightRequests.Inc()
		defer observability.HTTPInFlightRequests.Dec()

		// Track request size
		if r.ContentLength > 0 {
			observability.HTTPRequestSize.WithLabelValues(
				r.Method,
				sanitizedPath,
			).Observe(float64(r.ContentLength))
		}

		// Wrap ResponseWriter to capture status code and response size
		wrappedWriter := &prometheusResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			bytesWritten:   0,
		}

		next.ServeHTTP(wrappedWriter, r)

		// Calculate duration
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(wrappedWriter.statusCode)

		// Record metrics
		observability.HTTPRequestsTotal.WithLabelValues(
			r.Method,
			sanitizedPath,
			status,
		).Inc()

		observability.HTTPRequestDuration.WithLabelValues(
			r.Method,
			sanitizedPath,
			status,
		).Observe(duration)

		// Track response size
		if wrappedWriter.bytesWritten > 0 {
			observability.HTTPResponseSize.WithLabelValues(
				r.Method,
				sanitizedPath,
			).Observe(float64(wrappedWriter.bytesWritten))
		}
	})
}

// sanitizePathForPrometheus sanitizes URL paths for use in Prometheus labels
// It handles invalid UTF-8, percent-encoding, and limits path length
func sanitizePathForPrometheus(path string) string {
	// Handle empty path
	if path == "" {
		return "/"
	}

	// Decode percent-encoded paths
	decoded, err := url.PathUnescape(path)
	if err != nil {
		// If decoding fails, use original but sanitize it
		decoded = path
	}

	// Remove invalid UTF-8 characters
	if !utf8.ValidString(decoded) {
		// Convert to valid UTF-8 by replacing invalid bytes
		decoded = strings.ToValidUTF8(decoded, "_")
	}

	// Limit path length to prevent label cardinality explosion
	const maxPathLength = 100
	if len(decoded) > maxPathLength {
		decoded = decoded[:maxPathLength] + "..."
	}

	// Replace problematic characters for Prometheus labels
	// Keep alphanumeric, /, -, _, . and replace others
	var builder strings.Builder
	for _, r := range decoded {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '/' || r == '-' || r == '_' || r == '.' {
			builder.WriteRune(r)
		} else {
			builder.WriteRune('_')
		}
	}

	result := builder.String()

	// Ensure we don't return empty string
	if result == "" {
		return "/_invalid_"
	}

	return result
}

// prometheusResponseWriter wraps http.ResponseWriter to capture status code and bytes written
type prometheusResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (prw *prometheusResponseWriter) WriteHeader(code int) {
	prw.statusCode = code
	prw.ResponseWriter.WriteHeader(code)
}

func (prw *prometheusResponseWriter) Write(b []byte) (int, error) {
	n, err := prw.ResponseWriter.Write(b)
	prw.bytesWritten += n
	return n, err
}
