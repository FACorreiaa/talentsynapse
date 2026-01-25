package connections

import (
	"net/http"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/auth"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/matches"
	connectionspages "github.com/FACorreiaa/talentsynapse/internal/app/views/pages/connections"
)

// Handler handles connections HTTP requests
type Handler struct {
	matchesRepo *matches.Repository
}

// NewHandler creates a new connections handler
func NewHandler(repo *matches.Repository) *Handler {
	return &Handler{matchesRepo: repo}
}

// List renders the connections page with mutual matches
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get mutual matches (connections)
	connections, err := h.matchesRepo.GetMutualMatches(r.Context(), sessionData.UserID, 50)
	if err != nil {
		http.Error(w, "Failed to load connections", http.StatusInternalServerError)
		return
	}

	// Convert to view models
	var viewConnections []connectionspages.ConnectionCard
	for _, c := range connections {
		viewConnections = append(viewConnections, connectionspages.ConnectionCard{
			UserID:      c.UserID,
			DisplayName: c.DisplayName,
			Username:    c.Username,
			AvatarURL:   c.AvatarURL,
			Bio:         truncateBio(c.Bio, 100),
		})
	}

	component := connectionspages.Connections(
		viewConnections,
		sessionData.UserName,
		sessionData.UserAvatar,
	)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func truncateBio(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
