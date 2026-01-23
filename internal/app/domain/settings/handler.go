package settings

import (
	"net/http"

	"github.com/FACorreiaa/skillsphere/internal/app/domain/auth"
	settingspages "github.com/FACorreiaa/skillsphere/internal/app/views/pages/settings"
)

// Handler handles settings page HTTP requests
type Handler struct{}

// NewHandler creates a new settings handler
func NewHandler() *Handler {
	return &Handler{}
}

// Show renders the settings page
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	component := settingspages.Settings(
		sessionData.UserName,
		sessionData.UserAvatar,
		sessionData.UserEmail,
	)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
