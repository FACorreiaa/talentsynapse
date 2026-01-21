package auth

import (
	"context"
	"net/http"

	"github.com/FACorreiaa/skillsphere-pwa/internal/app/user"
)

// ContextKey is a type for context keys
type ContextKey string

const (
	SessionDataKey ContextKey = "session_data"
)

// RequireAuth middleware redirects unauthenticated users
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsAuthenticated(r) {
			session, _ := GetSession(r)
			session.Values["redirect_after_login"] = r.URL.Path
			session.Save(r, w)

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RedirectIfAuthenticated redirects logged-in users away from auth pages
func RedirectIfAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsAuthenticated(r) {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// InjectSessionData middleware adds session data to context
func InjectSessionData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionData := GetSessionData(r)
		ctx := context.WithValue(r.Context(), SessionDataKey, sessionData)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetSessionDataFromContext retrieves session data from context
func GetSessionDataFromContext(ctx context.Context) user.SessionData {
	if data, ok := ctx.Value(SessionDataKey).(user.SessionData); ok {
		return data
	}
	return user.SessionData{}
}
