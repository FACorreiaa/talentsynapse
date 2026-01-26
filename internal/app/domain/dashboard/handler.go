package dashboard

import (
	"context"
	"log"
	"net/http"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/auth"
	dashboardpages "github.com/FACorreiaa/talentsynapse/internal/app/views/pages/dashboard"
)

// UserStatsRepository provides access to user points/tier.
type UserStatsRepository interface {
	GetUserStats(ctx context.Context, userID string) (int, string, error)
}

// Handler handles dashboard HTTP requests
type Handler struct {
	userRepo UserStatsRepository
}

// NewHandler creates a new dashboard handler
func NewHandler(userRepo UserStatsRepository) *Handler {
	return &Handler{userRepo: userRepo}
}

// Show renders the dashboard page
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionDataFromContext(r.Context())

	flashes := auth.GetFlash(w, r)
	var successMsg string
	for _, flash := range flashes {
		if flash.Type == auth.FlashSuccess {
			successMsg = flash.Message
			break
		}
	}

	points := 0
	tier := "Bronze"
	if sessionData.UserID != "" && h.userRepo != nil {
		if p, t, err := h.userRepo.GetUserStats(r.Context(), sessionData.UserID); err == nil {
			points, tier = p, t
		} else {
			log.Printf("failed to load user stats for %s: %v", sessionData.UserID, err)
		}
	}

	component := dashboardpages.Dashboard(sessionData.UserName, sessionData.UserAvatar, successMsg, points, tier)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
