package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metrics for HTTP requests
var (
	// HTTPRequestsTotal counts total HTTP requests by method, path, and status
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration tracks request duration in seconds
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets, // 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestSize tracks request payload size in bytes
	HTTPRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000},
		},
		[]string{"method", "path"},
	)

	// HTTPResponseSize tracks response payload size in bytes
	HTTPResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000},
		},
		[]string{"method", "path"},
	)

	// ActiveUsers tracks currently active user sessions
	ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users_total",
			Help: "Number of currently active users",
		},
	)

	// HTTPInFlightRequests tracks concurrent requests being processed
	HTTPInFlightRequests = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_in_flight_requests",
			Help: "Number of HTTP requests currently being processed",
		},
	)
)
