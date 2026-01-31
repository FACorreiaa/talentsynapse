package auth

import (
	"context"
	"net/http"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/user"
)

// ContextKey is a type for context keys
type ContextKey string

const (
	SessionDataKey ContextKey = "session_data"
)

// RequireAuth middleware redirects unauthenticated users to login
// It validates:
// 1. Session cookie exists and contains a valid user ID
// 2. User ID is not empty
// 3. Stores the original URL for redirect after login
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if user is authenticated
		if !IsAuthenticated(r) {
			// Store the original URL for redirect after login
			session, err := GetSession(r)
			if err == nil {
				// Only save redirect for GET requests (not POST/DELETE etc.)
				if r.Method == http.MethodGet {
					session.Values["redirect_after_login"] = r.URL.Path
					if r.URL.RawQuery != "" {
						session.Values["redirect_after_login"] = r.URL.Path + "?" + r.URL.RawQuery
					}
					_ = session.Save(r, w)
				}
			}

			// For HTMX requests, return 401 with HX-Redirect header
			if isHTMXRequest(r) {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Additional validation: ensure session data is complete
		sessionData := GetSessionData(r)
		if sessionData.UserID == "" {
			// Session is corrupted, clear it and redirect to login
			_ = DestroySession(w, r)

			if isHTMXRequest(r) {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAuthWithUserCheck is an enhanced version that can validate user status
// against the database (for cases where user might be banned/deleted during session)
func RequireAuthWithUserCheck(userRepo UserStatusChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !IsAuthenticated(r) {
				handleUnauthenticated(w, r)
				return
			}

			sessionData := GetSessionData(r)
			if sessionData.UserID == "" {
				_ = DestroySession(w, r)
				handleUnauthenticated(w, r)
				return
			}

			// Optional: Check if user is still active in database
			if userRepo != nil {
				isActive, err := userRepo.IsUserActive(r.Context(), sessionData.UserID)
				if err != nil || !isActive {
					_ = DestroySession(w, r)
					_ = SetFlash(w, r, "Your session has expired. Please log in again.", FlashError)
					handleUnauthenticated(w, r)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// UserStatusChecker interface for checking user status
type UserStatusChecker interface {
	IsUserActive(ctx context.Context, userID string) (bool, error)
}

// handleUnauthenticated handles redirect for unauthenticated requests
func handleUnauthenticated(w http.ResponseWriter, r *http.Request) {
	if isHTMXRequest(r) {
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// isHTMXRequest checks if the request is an HTMX request
func isHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
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

// RequireAdmin middleware allows access only to admin users
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// First ensure user is authenticated
		if !IsAuthenticated(r) {
			session, _ := GetSession(r)
			session.Values["redirect_after_login"] = r.URL.Path
			_ = session.Save(r, w)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		sessionData := GetSessionData(r)
		if sessionData.Role != "admin" {
			// Render 403 Forbidden page or redirect
			// For now, redirect to dashboard with error flash
			_ = SetFlash(w, r, "Access denied. Admin privileges required.", FlashError)
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}
