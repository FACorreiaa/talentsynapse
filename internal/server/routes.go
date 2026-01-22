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

	"github.com/FACorreiaa/skillsphere/assets"
	"github.com/FACorreiaa/skillsphere/internal/app/auth"
	"github.com/FACorreiaa/skillsphere/internal/container"
)

// RegisterRoutes sets up all routes and middleware
func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()

	// Create dependency injection container
	c := container.New(s.db.GetPool())

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

	// Inject session data into all requests
	r.Use(auth.InjectSessionData)

	// ──────────────────────────────────────────────────────────────────
	// Security Middleware
	// ──────────────────────────────────────────────────────────────────

	// Secure Headers (HSTS, SSL Redirect, CSP, etc)
	isDev := os.Getenv("GO_ENV") == "development"
	secureOpts := secure.Options{
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://unpkg.com;",
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
	// Health Check
	// ──────────────────────────────────────────────────────────────────
	r.Get("/health", s.handleHealth)

	// ──────────────────────────────────────────────────────────────────
	// Public Routes
	// ──────────────────────────────────────────────────────────────────
	r.Get("/", c.HomeHandler.Index)

	// Auth routes (redirect if already authenticated)
	r.Group(func(r chi.Router) {
		r.Use(auth.RedirectIfAuthenticated)
		r.Get("/login", c.AuthHandler.ShowLogin)
		r.Post("/login", c.AuthHandler.HandleLogin)
		r.Get("/register", c.AuthHandler.ShowRegister)
		r.Post("/register", c.AuthHandler.HandleRegister)
		r.Get("/forgot-password", c.AuthHandler.ShowForgotPassword)
		r.Post("/forgot-password", c.AuthHandler.HandleForgotPassword)
	})

	// Logout (needs to work for authenticated users)
	r.Post("/logout", c.AuthHandler.HandleLogout)

	// ──────────────────────────────────────────────────────────────────
	// Protected Routes (require authentication)
	// ──────────────────────────────────────────────────────────────────
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth)
		r.Get("/dashboard", c.DashboardHandler.Show)

		// Profile routes
		r.Get("/profile", c.ProfileHandler.Show)
		r.Post("/profile", c.ProfileHandler.Update)

		// Skills routes
		r.Get("/skills", c.SkillsHandler.List)
		r.Get("/skills/add", c.SkillsHandler.ShowAdd)
		r.Post("/skills", c.SkillsHandler.Add)
		r.Delete("/skills/{id}", c.SkillsHandler.Remove)

		// Matches routes
		r.Get("/matches", c.MatchesHandler.List)
		r.Post("/matches/{id}/accept", c.MatchesHandler.Accept)
		r.Post("/matches/{id}/reject", c.MatchesHandler.Reject)

		// Settings routes
		r.Get("/settings", c.SettingsHandler.Show)
	})

	// Public discover page (works for both auth and non-auth)
	r.Get("/discover", c.DiscoverHandler.Show)

	// ──────────────────────────────────────────────────────────────────
	// API Routes
	// ──────────────────────────────────────────────────────────────────
	r.Route("/api", func(r chi.Router) {
		r.Get("/hello", s.handleAPIHello)
	})

	return r
}

// handleHealth returns service health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := s.db.Health()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(health)
}

// handleAPIHello is a sample JSON API endpoint
func (s *Server) handleAPIHello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Hello from SkillSphere!",
	})
}
