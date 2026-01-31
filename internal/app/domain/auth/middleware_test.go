package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/user"
)

// TestRequireAuth_UnauthenticatedUser tests that unauthenticated users are redirected
func TestRequireAuth_UnauthenticatedUser(t *testing.T) {
	// Initialize session store for testing
	require.NoError(t, os.Setenv("SESSION_SECRET", "test-secret-key-for-testing-purposes"))
	InitStore()

	tests := []struct {
		name             string
		path             string
		method           string
		isHTMX           bool
		expectedStatus   int
		expectedRedirect string
	}{
		{
			name:             "GET request redirects to login",
			path:             "/dashboard",
			method:           http.MethodGet,
			isHTMX:           false,
			expectedStatus:   http.StatusSeeOther,
			expectedRedirect: "/login",
		},
		{
			name:             "POST request redirects to login",
			path:             "/profile",
			method:           http.MethodPost,
			isHTMX:           false,
			expectedStatus:   http.StatusSeeOther,
			expectedRedirect: "/login",
		},
		{
			name:             "HTMX request returns 401 with HX-Redirect",
			path:             "/settings",
			method:           http.MethodGet,
			isHTMX:           true,
			expectedStatus:   http.StatusUnauthorized,
			expectedRedirect: "",
		},
		{
			name:             "Protected skills page redirects",
			path:             "/skills",
			method:           http.MethodGet,
			isHTMX:           false,
			expectedStatus:   http.StatusSeeOther,
			expectedRedirect: "/login",
		},
		{
			name:             "Protected matches page redirects",
			path:             "/matches",
			method:           http.MethodGet,
			isHTMX:           false,
			expectedStatus:   http.StatusSeeOther,
			expectedRedirect: "/login",
		},
		{
			name:             "Protected chat page redirects",
			path:             "/chat",
			method:           http.MethodGet,
			isHTMX:           false,
			expectedStatus:   http.StatusSeeOther,
			expectedRedirect: "/login",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler that should not be reached
			handler := RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("Handler should not be called for unauthenticated user")
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.isHTMX {
				req.Header.Set("HX-Request", "true")
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if tt.isHTMX {
				assert.Equal(t, tt.expectedStatus, rec.Code)
				assert.Equal(t, "/login", rec.Header().Get("HX-Redirect"))
			} else {
				assert.Equal(t, tt.expectedStatus, rec.Code)
				location := rec.Header().Get("Location")
				assert.Equal(t, tt.expectedRedirect, location)
			}
		})
	}
}

// TestRequireAuth_AuthenticatedUser tests that authenticated users can access protected routes
func TestRequireAuth_AuthenticatedUser(t *testing.T) {
	require.NoError(t, os.Setenv("SESSION_SECRET", "test-secret-key-for-testing-purposes"))
	InitStore()

	// Create a test user
	testUser := &user.User{
		ID:          "test-user-id-123",
		Email:       "test@example.com",
		Username:    "testuser",
		DisplayName: "Test User",
		Role:        "member",
	}

	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
	}{
		{
			name:           "authenticated user can access dashboard",
			path:           "/dashboard",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "authenticated user can access profile",
			path:           "/profile",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "authenticated user can access skills",
			path:           "/skills",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false
			handler := RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			// Create session for authenticated user
			err := CreateSession(rec, req, testUser)
			require.NoError(t, err)

			// Copy cookies from response to request
			for _, cookie := range rec.Result().Cookies() {
				req.AddCookie(cookie)
			}

			// Reset recorder
			rec = httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.True(t, handlerCalled, "Handler should be called for authenticated user")
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestRedirectIfAuthenticated_AuthenticatedUser tests redirect for logged-in users
func TestRedirectIfAuthenticated_AuthenticatedUser(t *testing.T) {
	require.NoError(t, os.Setenv("SESSION_SECRET", "test-secret-key-for-testing-purposes"))
	InitStore()

	testUser := &user.User{
		ID:          "test-user-id-123",
		Email:       "test@example.com",
		Username:    "testuser",
		DisplayName: "Test User",
		Role:        "member",
	}

	tests := []struct {
		name             string
		path             string
		expectedRedirect string
	}{
		{
			name:             "authenticated user redirected from login",
			path:             "/login",
			expectedRedirect: "/dashboard",
		},
		{
			name:             "authenticated user redirected from register",
			path:             "/register",
			expectedRedirect: "/dashboard",
		},
		{
			name:             "authenticated user redirected from forgot-password",
			path:             "/forgot-password",
			expectedRedirect: "/dashboard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := RedirectIfAuthenticated(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("Handler should not be called for authenticated user on auth pages")
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			// Create session
			err := CreateSession(rec, req, testUser)
			require.NoError(t, err)

			// Copy cookies
			for _, cookie := range rec.Result().Cookies() {
				req.AddCookie(cookie)
			}

			rec = httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusSeeOther, rec.Code)
			assert.Equal(t, tt.expectedRedirect, rec.Header().Get("Location"))
		})
	}
}

// TestRedirectIfAuthenticated_UnauthenticatedUser tests that unauth users can access auth pages
func TestRedirectIfAuthenticated_UnauthenticatedUser(t *testing.T) {
	require.NoError(t, os.Setenv("SESSION_SECRET", "test-secret-key-for-testing-purposes"))
	InitStore()

	tests := []struct {
		name string
		path string
	}{
		{name: "unauthenticated user can access login", path: "/login"},
		{name: "unauthenticated user can access register", path: "/register"},
		{name: "unauthenticated user can access forgot-password", path: "/forgot-password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false
			handler := RedirectIfAuthenticated(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.True(t, handlerCalled, "Handler should be called for unauthenticated user")
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

// TestRequireAdmin_NonAdminUser tests that non-admin users cannot access admin routes
func TestRequireAdmin_NonAdminUser(t *testing.T) {
	require.NoError(t, os.Setenv("SESSION_SECRET", "test-secret-key-for-testing-purposes"))
	InitStore()

	// Regular member user
	testUser := &user.User{
		ID:          "test-user-id-123",
		Email:       "test@example.com",
		Username:    "testuser",
		DisplayName: "Test User",
		Role:        "member",
	}

	handler := RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for non-admin user")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()

	// Create session
	err := CreateSession(rec, req, testUser)
	require.NoError(t, err)

	// Copy cookies
	for _, cookie := range rec.Result().Cookies() {
		req.AddCookie(cookie)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/dashboard", rec.Header().Get("Location"))
}

// TestRequireAdmin_AdminUser tests that admin users can access admin routes
func TestRequireAdmin_AdminUser(t *testing.T) {
	require.NoError(t, os.Setenv("SESSION_SECRET", "test-secret-key-for-testing-purposes"))
	InitStore()

	// Admin user
	testUser := &user.User{
		ID:          "admin-user-id-123",
		Email:       "admin@example.com",
		Username:    "adminuser",
		DisplayName: "Admin User",
		Role:        "admin",
	}

	handlerCalled := false
	handler := RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()

	// Create session
	err := CreateSession(rec, req, testUser)
	require.NoError(t, err)

	// Copy cookies
	for _, cookie := range rec.Result().Cookies() {
		req.AddCookie(cookie)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.True(t, handlerCalled, "Handler should be called for admin user")
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestRequireAdmin_UnauthenticatedUser tests that unauth users are redirected from admin
func TestRequireAdmin_UnauthenticatedUser(t *testing.T) {
	require.NoError(t, os.Setenv("SESSION_SECRET", "test-secret-key-for-testing-purposes"))
	InitStore()

	handler := RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for unauthenticated user")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/login", rec.Header().Get("Location"))
}

// TestIsHTMXRequest tests HTMX request detection
func TestIsHTMXRequest(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected bool
	}{
		{name: "HTMX request", header: "true", expected: true},
		{name: "non-HTMX request", header: "", expected: false},
		{name: "invalid HTMX header", header: "false", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.header != "" {
				req.Header.Set("HX-Request", tt.header)
			}
			assert.Equal(t, tt.expected, isHTMXRequest(req))
		})
	}
}

// TestRedirectAfterLogin tests that original URL is saved for redirect
func TestRedirectAfterLogin(t *testing.T) {
	require.NoError(t, os.Setenv("SESSION_SECRET", "test-secret-key-for-testing-purposes"))
	InitStore()

	tests := []struct {
		name         string
		originalPath string
		query        string
	}{
		{
			name:         "simple path is saved",
			originalPath: "/dashboard",
			query:        "",
		},
		{
			name:         "path with query is saved",
			originalPath: "/settings",
			query:        "tab=security",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			path := tt.originalPath
			if tt.query != "" {
				path = tt.originalPath + "?" + tt.query
			}

			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			// Verify redirect happened
			assert.Equal(t, http.StatusSeeOther, rec.Code)
			assert.Equal(t, "/login", rec.Header().Get("Location"))
		})
	}
}

// TestSessionDataInjection tests that session data is properly injected into context
func TestSessionDataInjection(t *testing.T) {
	require.NoError(t, os.Setenv("SESSION_SECRET", "test-secret-key-for-testing-purposes"))
	InitStore()

	testUser := &user.User{
		ID:          "test-user-id-123",
		Email:       "test@example.com",
		Username:    "testuser",
		DisplayName: "Test User",
		Role:        "member",
	}

	var capturedSessionData user.SessionData
	handler := InjectSessionData(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedSessionData = GetSessionDataFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Create session
	err := CreateSession(rec, req, testUser)
	require.NoError(t, err)

	// Copy cookies
	for _, cookie := range rec.Result().Cookies() {
		req.AddCookie(cookie)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, testUser.ID, capturedSessionData.UserID)
	assert.Equal(t, testUser.DisplayName, capturedSessionData.UserName)
	assert.Equal(t, testUser.Email, capturedSessionData.UserEmail)
	assert.Equal(t, testUser.Role, capturedSessionData.Role)
	assert.True(t, capturedSessionData.IsAuthenticated)
}

// MockUserStatusChecker implements UserStatusChecker for testing
type MockUserStatusChecker struct {
	isActive bool
	err      error
}

func (m *MockUserStatusChecker) IsUserActive(ctx context.Context, userID string) (bool, error) {
	return m.isActive, m.err
}

// TestRequireAuthWithUserCheck_InactiveUser tests that inactive users are logged out
func TestRequireAuthWithUserCheck_InactiveUser(t *testing.T) {
	require.NoError(t, os.Setenv("SESSION_SECRET", "test-secret-key-for-testing-purposes"))
	InitStore()

	testUser := &user.User{
		ID:          "test-user-id-123",
		Email:       "test@example.com",
		Username:    "testuser",
		DisplayName: "Test User",
		Role:        "member",
	}

	mockChecker := &MockUserStatusChecker{isActive: false}

	handler := RequireAuthWithUserCheck(mockChecker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for inactive user")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()

	// Create session
	err := CreateSession(rec, req, testUser)
	require.NoError(t, err)

	// Copy cookies
	for _, cookie := range rec.Result().Cookies() {
		req.AddCookie(cookie)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/login", rec.Header().Get("Location"))
}

// TestRequireAuthWithUserCheck_ActiveUser tests that active users can proceed
func TestRequireAuthWithUserCheck_ActiveUser(t *testing.T) {
	require.NoError(t, os.Setenv("SESSION_SECRET", "test-secret-key-for-testing-purposes"))
	InitStore()

	testUser := &user.User{
		ID:          "test-user-id-123",
		Email:       "test@example.com",
		Username:    "testuser",
		DisplayName: "Test User",
		Role:        "member",
	}

	mockChecker := &MockUserStatusChecker{isActive: true}

	handlerCalled := false
	handler := RequireAuthWithUserCheck(mockChecker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()

	// Create session
	err := CreateSession(rec, req, testUser)
	require.NoError(t, err)

	// Copy cookies
	for _, cookie := range rec.Result().Cookies() {
		req.AddCookie(cookie)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.True(t, handlerCalled, "Handler should be called for active user")
	assert.Equal(t, http.StatusOK, rec.Code)
}
