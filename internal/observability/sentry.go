package observability

import (
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
)

// InitSentry initializes the Sentry SDK for error tracking and performance monitoring
func InitSentry() {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		log.Println("⚠️  SENTRY_DSN not set, skipping Sentry initialization")
		return
	}

	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development"
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      env,
		TracesSampleRate: 1.0,
		EnableLogs:       true,
		Debug:            env == "development",
	}); err != nil {
		log.Printf("❌ Sentry initialization failed: %v\n", err)
		return
	}

	log.Println("✅ Sentry initialized successfully")
}

// FlushSentry waits for buffered events to be sent before shutdown
func FlushSentry() {
	sentry.Flush(2 * time.Second)
}

// TrackMetric sends custom metric data as Sentry breadcrumbs
// This allows you to track metrics in the context of transactions
func TrackMetric(name string, value float64, tags map[string]string) {
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Type:     "info",
		Category: "metric",
		Message:  name,
		Data: map[string]interface{}{
			"value": value,
			"tags":  tags,
		},
		Level:     sentry.LevelInfo,
		Timestamp: time.Now(),
	})
}

// IncrementCounter tracks counter metrics via breadcrumbs
func IncrementCounter(name string, tags map[string]string) {
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Type:     "info",
		Category: "counter",
		Message:  name,
		Data: map[string]interface{}{
			"tags": tags,
		},
		Level:     sentry.LevelInfo,
		Timestamp: time.Now(),
	})
}

// RecordGauge records gauge metrics via breadcrumbs
func RecordGauge(name string, value float64, tags map[string]string) {
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Type:     "info",
		Category: "gauge",
		Message:  name,
		Data: map[string]interface{}{
			"value": value,
			"tags":  tags,
		},
		Level:     sentry.LevelInfo,
		Timestamp: time.Now(),
	})
}
