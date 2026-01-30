package observability

import (
	"log"
	"os"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"
)

// NRApp holds the New Relic application instance
var NRApp *newrelic.Application

// InitNewRelic initializes the New Relic Go agent
func InitNewRelic() {
	licenseKey := os.Getenv("NEW_RELIC_LICENCE_KEY")
	appName := os.Getenv("NEW_RELIC_APP_NAME")

	if licenseKey == "" || appName == "" {
		log.Println("⚠️  New Relic not configured (missing NEW_RELIC_LICENCE_KEY or NEW_RELIC_APP_NAME)")
		return
	}

	region := os.Getenv("NEW_RELIC_REGION")

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(appName),
		newrelic.ConfigLicense(licenseKey),
		newrelic.ConfigDistributedTracerEnabled(true),
		newrelic.ConfigEnabled(true),
		// Set host based on region (EU vs US)
		func(cfg *newrelic.Config) {
			if region == "EU" {
				cfg.Host = "collector.eu01.nr-data.net"
			}
		},
		// Enable detailed logging in development
		newrelic.ConfigInfoLogger(os.Stdout),
	)
	if err != nil {
		log.Printf("❌ Failed to initialize New Relic: %v", err)
		return
	}

	// Wait for connection with timeout
	if err := app.WaitForConnection(5 * time.Second); err != nil {
		log.Printf("⚠️  New Relic connection timeout (will retry in background): %v", err)
	}

	NRApp = app
	log.Printf("✅ New Relic initialized for app: %s (region: %s)", appName, region)
}

// ShutdownNewRelic gracefully shuts down the New Relic agent
func ShutdownNewRelic() {
	if NRApp != nil {
		NRApp.Shutdown(10 * time.Second)
		log.Println("🛑 New Relic shutdown complete")
	}
}

// GetNewRelicApp returns the New Relic application instance
func GetNewRelicApp() *newrelic.Application {
	return NRApp
}
