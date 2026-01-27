// Package middleware provides HTTP middleware for the application.
package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/FACorreiaa/talentsynapse/internal/observability"
)

// PrometheusMiddleware tracks detailed HTTP metrics for Prometheus
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Track in-flight requests
		observability.HTTPInFlightRequests.Inc()
		defer observability.HTTPInFlightRequests.Dec()

		// Track request size
		if r.ContentLength > 0 {
			observability.HTTPRequestSize.WithLabelValues(
				r.Method,
				r.URL.Path,
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
			r.URL.Path,
			status,
		).Inc()

		observability.HTTPRequestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
			status,
		).Observe(duration)

		// Track response size
		if wrappedWriter.bytesWritten > 0 {
			observability.HTTPResponseSize.WithLabelValues(
				r.Method,
				r.URL.Path,
			).Observe(float64(wrappedWriter.bytesWritten))
		}
	})
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
