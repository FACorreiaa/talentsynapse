package dashboard

import (
	"net/http"

	"github.com/FACorreiaa/skillsphere/internal/app/domain/auth"
	dashboardpages "github.com/FACorreiaa/skillsphere/internal/app/views/pages/dashboard"
)

// Handler handles dashboard HTTP requests
type Handler struct{}

// NewHandler creates a new dashboard handler
func NewHandler() *Handler {
	return &Handler{}
}

// Show renders the dashboard page
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)

	flashes := auth.GetFlash(w, r)
	var successMsg string
	for _, flash := range flashes {
		if flash.Type == auth.FlashSuccess {
			successMsg = flash.Message
			break
		}
	}

	component := dashboardpages.Dashboard(sessionData.UserName, sessionData.UserAvatar, successMsg)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
