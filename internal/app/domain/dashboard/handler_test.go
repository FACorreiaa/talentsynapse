package dashboard

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/auth"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/user"
)

func attachSession(r *http.Request, data user.SessionData) *http.Request {
	ctx := context.WithValue(r.Context(), auth.SessionDataKey, data)
	return r.WithContext(ctx)
}

func TestShowRendersDefaultsWhenNoRepo(t *testing.T) {
	// Handler with nil repo should render defaults
	handler := NewHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req = attachSession(req, user.SessionData{UserID: "user-123", UserName: "Test User", UserAvatar: ""})
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	body := rr.Body.String()
	// Should render the welcome message with user name
	if !strings.Contains(body, "Test User") {
		t.Fatalf("expected user name to be rendered, got body: %s", body)
	}
	// Should render default tier
	if !strings.Contains(body, "Bronze") {
		t.Fatalf("expected fallback tier to be Bronze, got body: %s", body)
	}
	// Should render 0 points
	if !strings.Contains(body, ">0<") || !strings.Contains(body, "Points") {
		t.Fatalf("expected 0 points to be rendered, got body: %s", body)
	}
}

func TestShowRendersWithEmptySession(t *testing.T) {
	handler := NewHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	// Attach session with empty user data
	req = attachSession(req, user.SessionData{UserID: "", UserName: "Guest", UserAvatar: ""})
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	// Should return 200 OK
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}
