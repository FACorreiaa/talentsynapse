package matches

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/FACorreiaa/skillsphere/internal/app/domain/auth"
	matchespages "github.com/FACorreiaa/skillsphere/internal/app/views/pages/matches"
)

// Handler handles matches HTTP requests
type Handler struct {
	repo *Repository
}

// NewHandler creates a new matches handler
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// List renders the matches page with potential skill exchange partners
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get potential matches
	matches, err := h.repo.GetPendingMatches(r.Context(), sessionData.UserID, 20)
	if err != nil {
		http.Error(w, "Failed to load matches", http.StatusInternalServerError)
		return
	}

	// Convert to view models
	var matchCards []matchespages.MatchCard
	for _, m := range matches {
		var skills []matchespages.SkillOverlap
		for _, s := range m.OverlapSkills {
			skills = append(skills, matchespages.SkillOverlap{
				Name:      s.Name,
				TheyTeach: s.TheyTeach,
			})
		}

		matchCards = append(matchCards, matchespages.MatchCard{
			UserID:        m.UserID,
			DisplayName:   m.DisplayName,
			Username:      m.Username,
			AvatarURL:     m.AvatarURL,
			MatchScore:    int(m.MatchScore * 100),
			OverlapSkills: skills,
		})
	}

	flashes := auth.GetFlash(w, r)
	var successMsg string
	for _, flash := range flashes {
		if flash.Type == auth.FlashSuccess {
			successMsg = flash.Message
			break
		}
	}

	component := matchespages.Matches(
		matchCards,
		sessionData.UserName,
		sessionData.UserAvatar,
		successMsg,
	)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Accept handles accepting a match
func (h *Handler) Accept(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	matchedUserID := chi.URLParam(r, "id")
	if matchedUserID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	err := h.repo.RecordMatchAction(r.Context(), sessionData.UserID, matchedUserID, true)
	if err != nil {
		http.Error(w, "Failed to record match", http.StatusInternalServerError)
		return
	}

	// For HTMX requests, return empty to remove card
	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK)
		return
	}

	_ = auth.SetFlash(w, r, "Match accepted! You can now message this user.", auth.FlashSuccess)
	http.Redirect(w, r, "/matches", http.StatusSeeOther)
}

// Reject handles rejecting a match
func (h *Handler) Reject(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	matchedUserID := chi.URLParam(r, "id")
	if matchedUserID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	err := h.repo.RecordMatchAction(r.Context(), sessionData.UserID, matchedUserID, false)
	if err != nil {
		http.Error(w, "Failed to record rejection", http.StatusInternalServerError)
		return
	}

	// For HTMX requests, return empty to remove card
	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/matches", http.StatusSeeOther)
}
