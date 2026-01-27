package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/FACorreiaa/talentsynapse/internal/observability"
	"github.com/FACorreiaa/talentsynapse/internal/server"
)

func main() {
	// Initialize observability
	observability.InitSentry()
	defer observability.FlushSentry()

	// Create server instance
	s := server.New()

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Determine deployment mode
	domain := os.Getenv("DOMAIN")  // e.g., "TalentSynapse.com"
	useTLS := os.Getenv("USE_TLS") // "true" for built-in TLS
	goEnv := os.Getenv("GO_ENV")   // "production" or "development"

	// Start server based on configuration
	go func() {
		if goEnv == "production" && useTLS == "true" && domain != "" {
			// Path A: Pure Go with built-in Let's Encrypt
			fmt.Printf("🔒 Starting with built-in TLS for domain: %s\n", domain)
			fmt.Println("⚠️  Note: Binary must run as root or have CAP_NET_BIND_SERVICE")
			fmt.Println("⚠️  Ensure port 80 and 443 are open and DNS is configured")

			// Use the built-in TLS server (this will block)
			if err := s.ListenAndServeTLS(domain); err != nil {
				log.Fatalf("TLS Server error: %v", err)
			}
		} else {
			// Path B: Standard HTTP (use Caddy/nginx for TLS in production)
			port := os.Getenv("PORT")
			if port == "" {
				port = "8080"
			}

			if goEnv == "production" {
				fmt.Printf("🚀 Server starting on :%s (use Caddy/nginx for TLS)\n", port)
			} else {
				fmt.Printf("🚀 Server starting on http://localhost:%s\n", port)
			}

			srv := s.HTTPServer()
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("Server error: %v", err)
			}
		}
	}()

	<-done
	log.Println("Server stopped gracefully")
}
