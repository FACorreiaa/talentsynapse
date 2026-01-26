package dashboard

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/auth"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/user"
)

type mockUserStatsRepo struct {
	points int
	tier   string
	err    error
}

func (m *mockUserStatsRepo) GetUserStats(ctx context.Context, userID string) (int, string, error) {
	return m.points, m.tier, m.err
}

func attachSession(r *http.Request, data user.SessionData) *http.Request {
	ctx := context.WithValue(r.Context(), auth.SessionDataKey, data)
	return r.WithContext(ctx)
}

func TestShowRendersUserStats(t *testing.T) {
	repo := &mockUserStatsRepo{points: 120, tier: "Silver"}
	handler := NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req = attachSession(req, user.SessionData{UserID: "user-123", UserName: "Test User", UserAvatar: ""})
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "120") {
		t.Fatalf("expected points to be rendered, got body: %s", body)
	}
	if !strings.Contains(body, "Silver") {
		t.Fatalf("expected tier to be rendered, got body: %s", body)
	}
}

func TestShowFallsBackOnStatsError(t *testing.T) {
	repo := &mockUserStatsRepo{err: errors.New("boom")}
	handler := NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req = attachSession(req, user.SessionData{UserID: "user-123", UserName: "Test User", UserAvatar: ""})
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "Bronze") {
		t.Fatalf("expected fallback tier to be Bronze, got body: %s", body)
	}
}
