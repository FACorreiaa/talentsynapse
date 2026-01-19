package middleware

import (
	"net/http"

	"github.com/unrolled/secure"
)

// SecureHeaders returns middleware that sets security headers
func SecureHeaders(allowedHosts []string, isDevelopment bool) func(next http.Handler) http.Handler {
	secureMiddleware := secure.New(secure.Options{
		// Allowed hosts
		AllowedHosts:         allowedHosts,
		AllowedHostsAreRegex: false,
		HostsProxyHeaders:    []string{"X-Forwarded-Host"},

		// SSL/HTTPS
		SSLRedirect:     !isDevelopment,
		SSLHost:         "",
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},

		// HSTS (HTTP Strict Transport Security)
		STSSeconds:           31536000, // 1 year
		STSIncludeSubdomains: true,
		STSPreload:           true,

		// Frame options
		FrameDeny: true,
		// Or use: CustomFrameOptionsValue: "SAMEORIGIN"

		// Content type options
		ContentTypeNosniff: true,

		// XSS Protection
		BrowserXssFilter: true,

		// Content Security Policy
		ContentSecurityPolicy: "default-src 'self'; " +
			"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; " +
			"script-src 'self' 'unsafe-inline' https://unpkg.com; " +
			"img-src 'self' data: https:; " +
			"font-src 'self' https://fonts.gstatic.com; " +
			"connect-src 'self';",

		// Referrer policy
		ReferrerPolicy: "strict-origin-when-cross-origin",

		// Development mode (disables some checks)
		IsDevelopment: isDevelopment,
	})

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			secureMiddleware.Handler(next).ServeHTTP(w, r)
		})
	}
}

// DefaultSecureHeaders returns secure headers middleware with sensible defaults
func DefaultSecureHeaders(isDevelopment bool) func(next http.Handler) http.Handler {
	return SecureHeaders([]string{"localhost", "127.0.0.1"}, isDevelopment)
}
