package session

import (
	"net/http"
	"time"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/auth"
	sessionpages "github.com/FACorreiaa/talentsynapse/internal/app/views/pages/sessions"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// List shows the "My Sessions" page
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	sessions, err := h.repo.GetUserSessions(r.Context(), sessionData.UserID)
	if err != nil {
		http.Error(w, "Failed to load sessions", http.StatusInternalServerError)
		return
	}

	// Map to view model
	var viewSessions []sessionpages.SessionItem
	for _, s := range sessions {
		partnerName := s.PartnerName
		partnerAvatar := s.PartnerAvatar
		partnerID := s.PartnerID
		if s.InitiatorID != sessionData.UserID {
			partnerName = s.InitiatorName
			partnerAvatar = s.InitiatorAvatar
			partnerID = s.InitiatorID
		}

		viewSessions = append(viewSessions, sessionpages.SessionItem{
			ID:             s.ID,
			PartnerID:      partnerID,
			PartnerName:    partnerName,
			PartnerAvatar:  partnerAvatar,
			Status:         s.Status,
			ScheduledStart: s.ScheduledStart,
			IsReviewed:     s.IsReviewed,
		})
	}

	component := sessionpages.List(viewSessions, sessionData.UserName, sessionData.UserAvatar)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// CreateRequest handles creating a new quick session (from profile)
func (h *Handler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	partnerID := r.FormValue("partner_id")

	// Default to "Now"
	startTime := time.Now()

	if err := h.repo.CreateRequest(r.Context(), sessionData.UserID, partnerID, startTime); err != nil {
		_ = auth.SetFlash(w, r, "Failed to request session", auth.FlashError)
	} else {
		_ = auth.SetFlash(w, r, "Session requested", auth.FlashSuccess)
	}

	http.Redirect(w, r, "/sessions", http.StatusSeeOther)
}

// Complete handles marking a session as completed
func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := chi.URLParam(r, "id")
	if err := h.repo.MarkCompleted(r.Context(), sessionID, sessionData.UserID); err != nil {
		_ = auth.SetFlash(w, r, "Failed to complete session", auth.FlashError)
	} else {
		_ = auth.SetFlash(w, r, "Session completed! Please delete a review.", auth.FlashSuccess)
	}

	http.Redirect(w, r, "/sessions", http.StatusSeeOther)
}
