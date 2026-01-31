package session

import (
	"log"
	"net/http"
	"time"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/auth"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/badges"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/matches"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/points"
	sessionpages "github.com/FACorreiaa/talentsynapse/internal/app/views/pages/sessions"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	repo          *Repository
	badgesRepo    *badges.Repository
	matchesRepo   *matches.Repository
	pointsService *points.Service
}

func NewHandler(repo *Repository, badgesRepo *badges.Repository, matchesRepo *matches.Repository, pointsService *points.Service) *Handler {
	return &Handler{
		repo:          repo,
		badgesRepo:    badgesRepo,
		matchesRepo:   matchesRepo,
		pointsService: pointsService,
	}
}

// List shows the "My Sessions" page
// Redirects to matches page if user has no confirmed connections
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Check if user has any confirmed connections
	hasConnections, err := h.matchesRepo.HasConnections(r.Context(), sessionData.UserID)
	if err != nil {
		// Log error but continue - don't block the page
		hasConnections = false
	}

	// If no connections, redirect to matches page to find people
	if !hasConnections {
		_ = auth.SetFlash(w, r, "Find and connect with people first to schedule sessions!", auth.FlashInfo)
		http.Redirect(w, r, "/matches", http.StatusSeeOther)
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
		_ = auth.SetFlash(w, r, "Session completed! Please leave a review.", auth.FlashSuccess)

		userUUID := uuid.MustParse(sessionData.UserID)

		// Award points for completing session
		if h.pointsService != nil {
			upgrade, err := h.pointsService.AwardSessionPoints(r.Context(), userUUID)
			if err != nil {
				log.Printf("Failed to award session points: %v", err)
			} else if upgrade != nil {
				log.Printf("User %s upgraded to %s tier!", sessionData.UserID, upgrade.NewTier)
			}
		}

		// Gamification: Check for dedicated_learner badge (5+ completed sessions)
		count, err := h.repo.CountCompletedSessions(r.Context(), sessionData.UserID)
		if err == nil && count >= 5 {
			_ = h.badgesRepo.AwardBadge(r.Context(), userUUID, "dedicated_learner")
		}

		// Award first session badge if this is the first
		if err == nil && count == 1 && h.pointsService != nil {
			_, _ = h.pointsService.AwardFirstSessionPoints(r.Context(), userUUID)
		}
	}

	http.Redirect(w, r, "/sessions", http.StatusSeeOther)
}
