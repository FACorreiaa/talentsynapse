// Package middleware provides HTTP middleware for the application.
// This file serves as a barrel export and documents available middleware.
package middleware

// Available Middleware (from chi and third-party):
// ─────────────────────────────────────────────────────────────────────
// Core (chi/middleware):
//   - RequestID    - Injects a request ID into the context
//   - RealIP       - Sets RemoteAddr from X-Real-IP or X-Forwarded-For
//   - Logger       - Logs the start and end of each request
//   - Recoverer    - Gracefully recovers from panics
//   - Timeout      - Signals to the request context when timeout is reached
//   - Compress     - Gzip compression for responses
//   - StripSlashes - Strips trailing slashes from requests
//   - RedirectSlashes - Redirects trailing slash requests
//   - Heartbeat    - Handles heartbeat/ping requests
//
// Security:
//   - Secure       - Security headers (CSP, HSTS, etc) via github.com/unrolled/secure
//   - CORS         - Cross-Origin Resource Sharing via github.com/go-chi/cors
//   - RateLimiter  - Rate limiting via github.com/go-chi/httprate
//
// TODO: Implement these middleware as needed:
// ─────────────────────────────────────────────────────────────────────

// BasicAuth - HTTP Basic Authentication
// See: https://github.com/go-chi/chi/blob/master/middleware/basic_auth.go

// JWT - JSON Web Token authentication
// Recommended: github.com/golang-jwt/jwt/v5
// Example implementation in jwt.go.example

// Session - Session management
// Recommended: github.com/gorilla/sessions or github.com/alexedwards/scs/v2

// CSRF - Cross-Site Request Forgery protection
// Recommended: github.com/gorilla/csrf or github.com/justinas/nosurf

// KeyAuth - API Key authentication
// Custom implementation required

// Jaeger - Distributed tracing
// See: go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp

// Prometheus - Metrics collection
// See: github.com/prometheus/client_golang/prometheus/promhttp

// Proxy - Reverse proxy support
// See: net/http/httputil.ReverseProxy

// Rewrite - URL rewriting
// Custom implementation required
