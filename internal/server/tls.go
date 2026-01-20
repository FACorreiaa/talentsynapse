package server

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

// ListenAndServeTLS starts the server with automatic HTTPS (Let's Encrypt)
// This is an ALTERNATIVE to using Caddy - only use if you want pure Go deployment
func (s *Server) ListenAndServeTLS(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain is required for TLS")
	}

	// Let's Encrypt certificate manager
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
		Cache:      autocert.DirCache("/var/www/.cache"), // Must be writable
		Email:      os.Getenv("ADMIN_EMAIL"),             // Optional but recommended
	}

	// HTTPS server on :443
	httpsServer := &http.Server{
		Addr:    ":443",
		Handler: s.RegisterRoutes(),
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			},
		},
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// HTTP server on :80 (for ACME challenge and redirect)
	httpServer := &http.Server{
		Addr:         ":80",
		Handler:      certManager.HTTPHandler(http.HandlerFunc(redirectToHTTPS)),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start HTTP server in background
	go func() {
		log.Println("🔓 HTTP server starting on :80 (ACME challenge + redirect)")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Start HTTPS server (blocks)
	log.Printf("🔒 HTTPS server starting on :443 (domain: %s)", domain)
	log.Println("⚠️  Note: First request will be slow while obtaining certificate")
	return httpsServer.ListenAndServeTLS("", "")
}

// redirectToHTTPS redirects all HTTP traffic to HTTPS
func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	target := "https://" + r.Host + r.URL.Path
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target, http.StatusMovedPermanently)
}
