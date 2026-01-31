package server

import (
	"encoding/json"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/unrolled/secure"

	"github.com/FACorreiaa/talentsynapse/assets"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/auth"
	"github.com/FACorreiaa/talentsynapse/internal/container"
	customMiddleware "github.com/FACorreiaa/talentsynapse/internal/middleware"
)

// RegisterRoutes sets up all routes and middleware
func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()

	// Create dependency injection container
	c := container.New(s.db.GetPool())

	// ──────────────────────────────────────────────────────────────────
	// Core Middleware
	// ──────────────────────────────────────────────────────────────────

	// New Relic APM middleware (should be very early to capture full transaction)
	r.Use(customMiddleware.NewRelicMiddleware)

	// Sentry middleware (should be early in the chain to catch all errors)
	sentryHandler := sentryhttp.New(sentryhttp.Options{
		Repanic:         true,
		WaitForDelivery: false,
		Timeout:         2 * time.Second,
	})
	r.Use(func(next http.Handler) http.Handler {
		return sentryHandler.Handle(next)
	})

	// Prometheus metrics middleware (detailed HTTP metrics)
	r.Use(customMiddleware.PrometheusMiddleware)

	// Sentry metrics middleware (breadcrumbs for context)
	r.Use(customMiddleware.MetricsMiddleware)

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
		// Allow localhost for health checks
		secureOpts.AllowedHosts = []string{"talentsynapse.org", "www.talentsynapse.org", "localhost:8080", "localhost"}
		secureOpts.SSLRedirect = true
		secureOpts.STSSeconds = 31536000
		secureOpts.STSIncludeSubdomains = true
		secureOpts.STSPreload = true
		secureOpts.SSLProxyHeaders = map[string]string{"X-Forwarded-Proto": "https"}
	}
	secureMiddleware := secure.New(secureOpts)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Skip SSL redirect for health check and metrics endpoints
			// These are called internally and don't need HTTPS
			if req.URL.Path == "/health" || req.URL.Path == "/metrics" {
				next.ServeHTTP(w, req)
				return
			}
			secureMiddleware.Handler(next).ServeHTTP(w, req)
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

	// Favicon - serve SVG icon (browsers support SVG favicons)
	r.Get("/favicon.ico", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/assets/static/icon.svg", http.StatusMovedPermanently)
	})
	r.Get("/favicon.svg", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/assets/static/icon.svg", http.StatusMovedPermanently)
	})

	// ──────────────────────────────────────────────────────────────────
	// Health Check & Metrics
	// ──────────────────────────────────────────────────────────────────
	r.Get("/health", s.handleHealth)

	// Prometheus metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

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

	// Email verification (works for both auth and non-auth users)
	r.Get("/verify-email", c.AuthHandler.HandleEmailVerification)
	r.Post("/resend-verification", c.AuthHandler.HandleResendVerification)

	// Password reset (public routes)
	r.Get("/reset-password", c.AuthHandler.ShowResetPassword)
	r.Post("/reset-password", c.AuthHandler.HandleResetPassword)

	// MFA routes (authentication flow)
	r.Get("/mfa/verify", c.AuthHandler.HandleMFAVerifyPage)
	r.Post("/mfa/verify", c.AuthHandler.HandleMFAVerifyCode)
	r.Post("/mfa/verify-backup", c.AuthHandler.HandleMFAVerifyBackupCode)

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
		r.Post("/profile/portfolio", c.PortfolioHandler.Create)
		r.Delete("/profile/portfolio/{id}", c.PortfolioHandler.Delete)

		// Public profile (view other users)
		r.Get("/users/{id}", c.ProfileHandler.ShowPublic)

		// Skills routes
		r.Get("/skills", c.SkillsHandler.List)
		r.Get("/skills/add", c.SkillsHandler.ShowAdd)
		r.Post("/skills", c.SkillsHandler.Add)
		r.Delete("/skills/{id}", c.SkillsHandler.Remove)

		// Matches routes
		r.Get("/matches", c.MatchesHandler.List)
		r.Post("/matches/{id}/accept", c.MatchesHandler.Accept)
		r.Post("/matches/{id}/reject", c.MatchesHandler.Reject)

		// Connections routes
		r.Get("/connections", c.ConnectionsHandler.List)

		// Chat routes
		r.Get("/chat", c.ChatHandler.ListConversations)
		r.Get("/chat/{id}", c.ChatHandler.ShowChat)
		r.Post("/chat/{id}/messages", c.ChatHandler.SendMessage)
		r.Get("/chat/start/{userID}", c.ChatHandler.StartConversation)
		r.Get("/chat/ws", c.ChatHub.HandleWebSocket)

		// Calendar & Scheduling routes
		r.Get("/calendar", c.SchedulingHandler.ShowCalendar)
		r.Get("/calendar/month", c.SchedulingHandler.GetCalendarMonth)
		r.Get("/calendar/session/new", c.SchedulingHandler.ProposeSessionForm)
		r.Post("/calendar/session/create", c.SchedulingHandler.CreateSession)
		r.Get("/calendar/session/{id}", c.SchedulingHandler.GetSessionDetails)
		r.Post("/calendar/session/{id}/confirm", c.SchedulingHandler.ConfirmSession)
		r.Post("/calendar/session/{id}/cancel", c.SchedulingHandler.CancelSession)
		r.Get("/calendar/ws", c.SchedulingHandler.HandleWebSocket)

		// Settings routes
		r.Get("/settings", c.SettingsHandler.Show)
		r.Get("/settings/tab/{tab}", c.SettingsHandler.Tab)
		r.Get("/settings/security", c.AuthHandler.HandleSecuritySettings)

		// MFA management routes (protected)
		r.Get("/mfa/setup", c.AuthHandler.HandleMFASetup)
		r.Post("/mfa/verify-setup", c.AuthHandler.HandleMFAVerifySetup)
		r.Post("/mfa/disable", c.AuthHandler.HandleMFADisable)
		r.Post("/mfa/regenerate-backup-codes", c.AuthHandler.HandleMFARegenerateBackupCodes)
		r.Post("/mfa/remove-trusted-device/*", c.AuthHandler.HandleRemoveTrustedDevice)
		r.Post("/mfa/remove-all-trusted-devices", c.AuthHandler.HandleRemoveAllTrustedDevices)

		// Push Notifications
		r.Get("/api/push/vapid-key", c.PushHandler.GetVAPIDKey)
		r.Post("/api/push/subscribe", c.PushHandler.Subscribe)

		// Report routes
		r.Post("/report", c.ReportHandler.SubmitReport)

		// Review routes
		r.Post("/review", c.ReviewHandler.Submit)

		// Session routes
		r.Get("/sessions", c.SessionHandler.List)
		r.Post("/sessions/request", c.SessionHandler.CreateRequest)
		r.Post("/sessions/{id}/complete", c.SessionHandler.Complete)

		// Admin routes
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAdmin)
			r.Get("/admin", c.AdminHandler.Dashboard)
			r.Post("/admin/users/{userID}/ban", c.AdminHandler.ToggleBan)
			r.Post("/admin/users/{userID}/verify", c.AdminHandler.ToggleVerify)
			r.Get("/admin/moderation", c.AdminHandler.ModerationQueue)
			r.Post("/admin/reports/{reportID}/resolve", c.AdminHandler.ResolveReport)
		})
	})

	// Public discover page (works for both auth and non-auth)
	r.Get("/discover", c.DiscoverHandler.Show)

	// ──────────────────────────────────────────────────────────────────
	// API Routes
	// ──────────────────────────────────────────────────────────────────
	r.Route("/api", func(r chi.Router) {
		r.Get("/hello", s.handleAPIHello)
	})

	// ──────────────────────────────────────────────────────────────────
	// Debug/Profiling Routes (pprof)
	// ──────────────────────────────────────────────────────────────────
	// Production-safe: requires authentication and can be restricted to localhost
	r.Route("/debug/pprof", func(r chi.Router) {
		r.Use(pprofAuthMiddleware)
		r.HandleFunc("/", pprof.Index)
		r.HandleFunc("/cmdline", pprof.Cmdline)
		r.HandleFunc("/profile", pprof.Profile)
		r.HandleFunc("/symbol", pprof.Symbol)
		r.HandleFunc("/trace", pprof.Trace)
		r.Handle("/goroutine", pprof.Handler("goroutine"))
		r.Handle("/heap", pprof.Handler("heap"))
		r.Handle("/threadcreate", pprof.Handler("threadcreate"))
		r.Handle("/block", pprof.Handler("block"))
		r.Handle("/mutex", pprof.Handler("mutex"))
		r.Handle("/allocs", pprof.Handler("allocs"))
	})

	// ──────────────────────────────────────────────────────────────────
	// 404 Not Found Handler
	// ──────────────────────────────────────────────────────────────────
	r.NotFound(c.ErrorHandler.NotFound)

	return r
}

// handleHealth returns service health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := s.db.Health()
	w.Header().Set("Content-Type", "application/json")

	// Return 503 if unhealthy
	if health["status"] == "unhealthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	_ = json.NewEncoder(w).Encode(health)
}

// handleAPIHello is a sample JSON API endpoint
func (s *Server) handleAPIHello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Hello from TalentSynapse!",
	})
}
