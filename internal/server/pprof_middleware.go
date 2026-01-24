package server

import (
	"crypto/subtle"
	"net/http"
	"os"
)

// pprofAuthMiddleware protects pprof endpoints with basic authentication in production
func pprofAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// In development, allow unrestricted access
		if os.Getenv("GO_ENV") == "development" {
			next.ServeHTTP(w, r)
			return
		}

		// In production, require authentication
		username := os.Getenv("PPROF_USERNAME")
		password := os.Getenv("PPROF_PASSWORD")

		// If credentials are not set, disable pprof entirely in production
		if username == "" || password == "" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		// Get credentials from Authorization header
		user, pass, ok := r.BasicAuth()

		// Use constant-time comparison to prevent timing attacks
		usernameMatch := subtle.ConstantTimeCompare([]byte(user), []byte(username)) == 1
		passwordMatch := subtle.ConstantTimeCompare([]byte(pass), []byte(password)) == 1

		if !ok || !usernameMatch || !passwordMatch {
			w.Header().Set("WWW-Authenticate", `Basic realm="pprof"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
