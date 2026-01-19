package server

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/unrolled/secure"

	"github.com/FACorreiaa/skillsphere-pwa/assets"
	"github.com/FACorreiaa/skillsphere-pwa/views/pages"
)

// RegisterRoutes sets up all routes and middleware
func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()

	// ──────────────────────────────────────────────────────────────────
	// Core Middleware
	// ──────────────────────────────────────────────────────────────────
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Context Timeout: cancels context if request takes > 60s
	r.Use(middleware.Timeout(60 * time.Second))

	// Compression (Gzip/Deflate)
	r.Use(middleware.Compress(5))

	// ──────────────────────────────────────────────────────────────────
	// Security Middleware
	// ──────────────────────────────────────────────────────────────────

	// Secure Headers (HSTS, SSL Redirect, CSP, etc)
	isDev := os.Getenv("GO_ENV") == "development"
	secureOpts := secure.Options{
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; script-src 'self' 'unsafe-inline' https://unpkg.com;",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}
	if !isDev {
		// Production-only strict settings
		secureOpts.AllowedHosts = []string{"yourdomain.com"}
		secureOpts.SSLRedirect = true
		secureOpts.STSSeconds = 31536000
		secureOpts.STSIncludeSubdomains = true
		secureOpts.STSPreload = true
		secureOpts.SSLProxyHeaders = map[string]string{"X-Forwarded-Proto": "https"}
	}
	secureMiddleware := secure.New(secureOpts)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			secureMiddleware.Handler(next).ServeHTTP(w, r)
		})
	})

	// CORS (Cross Origin Resource Sharing)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate Limiting (100 requests / 1 minute per IP)
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	// ──────────────────────────────────────────────────────────────────
	// Static Assets
	// ──────────────────────────────────────────────────────────────────
	if os.Getenv("GO_ENV") == "development" {
		// DEV: Serve from disk for hot reload
		fs := http.FileServer(http.Dir("./assets"))
		r.Handle("/assets/*", http.StripPrefix("/assets", fs))
	} else {
		// PROD: Serve from embedded binary
		fs := http.FileServer(http.FS(assets.Files))
		r.Handle("/assets/*", http.StripPrefix("/assets", fs))
	}

	// ──────────────────────────────────────────────────────────────────
	// Application Routes
	// ──────────────────────────────────────────────────────────────────

	// Health check
	r.Get("/health", s.handleHealth)

	// Pages
	r.Get("/", s.handleHome)

	// API routes (example)
	r.Route("/api", func(r chi.Router) {
		r.Get("/hello", s.handleAPIHello)
	})

	return r
}

// handleHealth returns service health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := s.db.Health()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// handleHome renders the home page using Templ
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	component := pages.Index()
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleAPIHello is a sample JSON API endpoint
func (s *Server) handleAPIHello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Hello from GoForge!",
	})
}
