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
