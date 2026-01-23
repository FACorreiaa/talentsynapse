package home

import (
	"net/http"

	"github.com/FACorreiaa/skillsphere/internal/app/domain/auth"
	"github.com/FACorreiaa/skillsphere/internal/app/views/pages"
)

// Handler handles home page HTTP requests
type Handler struct{}

// NewHandler creates a new home handler
func NewHandler() *Handler {
	return &Handler{}
}

// Index renders the landing page
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	component := pages.Index(sessionData.IsAuthenticated, sessionData.UserName)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
