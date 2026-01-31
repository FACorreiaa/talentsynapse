package matches

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/auth"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/badges"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/review"
	matchespages "github.com/FACorreiaa/talentsynapse/internal/app/views/pages/matches"
	"github.com/google/uuid"
)

// Handler handles matches HTTP requests
type Handler struct {
	repo         *Repository
	badgeService *badges.Service
	reviewRepo   *review.Repository
}

// NewHandler creates a new matches handler
func NewHandler(repo *Repository, badgeService *badges.Service, reviewRepo *review.Repository) *Handler {
	return &Handler{
		repo:         repo,
		badgeService: badgeService,
		reviewRepo:   reviewRepo,
	}
}

// List renders the matches page with potential skill exchange partners
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Check which tab is active (default to "discover" unless user has no matches but has connections)
	activeTab := r.URL.Query().Get("tab")
	if activeTab == "" {
		activeTab = "discover"
	}

	// Get mutual matches (connections)
	mutualMatches, err := h.repo.GetMutualMatches(r.Context(), sessionData.UserID, 50)
	if err != nil {
		mutualMatches = nil
	}

	// Convert mutual matches to connection cards
	var connectionCards []matchespages.ConnectionCard
	for _, m := range mutualMatches {
		connectionCards = append(connectionCards, matchespages.ConnectionCard{
			UserID:      m.UserID,
			DisplayName: m.DisplayName,
			Username:    m.Username,
			AvatarURL:   m.AvatarURL,
			Bio:         m.Bio,
		})
	}

	// Get pending sent matches (users YOU liked but who haven't responded)
	pendingSent, err := h.repo.GetPendingSentMatches(r.Context(), sessionData.UserID, 20)
	if err != nil {
		pendingSent = nil
	}

	// Convert pending sent to pending cards
	var pendingCards []matchespages.PendingCard
	for _, ps := range pendingSent {
		pendingCards = append(pendingCards, matchespages.PendingCard{
			UserID:      ps.UserID,
			DisplayName: ps.DisplayName,
			Username:    ps.Username,
			AvatarURL:   ps.AvatarURL,
		})
	}

	// Get pending received matches (users who liked you)
	pendingReceived, err := h.repo.GetPendingReceivedMatches(r.Context(), sessionData.UserID, 20)
	if err != nil {
		// Log error but continue
		pendingReceived = nil
	}

	// Convert pending received to match cards (prioritize these)
	var matchCards []matchespages.MatchCard
	for _, pr := range pendingReceived {
		// Get overlap skills for this user
		overlapSkills, _ := h.repo.GetOverlapSkills(r.Context(), sessionData.UserID, pr.UserID)
		var skills []matchespages.SkillOverlap
		for _, s := range overlapSkills {
			skills = append(skills, matchespages.SkillOverlap{
				Name:      s.Name,
				TheyTeach: s.TheyTeach,
			})
		}

		// Get rating stats
		reviewCount, avgRating, _ := h.reviewRepo.GetStats(r.Context(), uuid.MustParse(pr.UserID))

		matchCards = append(matchCards, matchespages.MatchCard{
			UserID:        pr.UserID,
			DisplayName:   pr.DisplayName,
			Username:      pr.Username,
			AvatarURL:     pr.AvatarURL,
			MatchScore:    int(pr.MatchScore * 100),
			OverlapSkills: skills,
			IsLikedYou:    true, // Flag to show "Liked You" badge
			ReviewCount:   reviewCount,
			AverageRating: avgRating,
		})
	}

	// Get potential matches (new suggestions)
	matches, err := h.repo.GetPendingMatches(r.Context(), sessionData.UserID, 20)
	if err != nil {
		http.Error(w, "Failed to load matches", http.StatusInternalServerError)
		return
	}

	// Convert to view models
	for _, m := range matches {
		var skills []matchespages.SkillOverlap
		for _, s := range m.OverlapSkills {
			skills = append(skills, matchespages.SkillOverlap{
				Name:      s.Name,
				TheyTeach: s.TheyTeach,
			})
		}

		// Get rating stats
		reviewCount, avgRating, _ := h.reviewRepo.GetStats(r.Context(), uuid.MustParse(m.UserID))

		matchCards = append(matchCards, matchespages.MatchCard{
			UserID:        m.UserID,
			DisplayName:   m.DisplayName,
			Username:      m.Username,
			AvatarURL:     m.AvatarURL,
			MatchScore:    int(m.MatchScore * 100),
			OverlapSkills: skills,
			IsLikedYou:    false,
			ReviewCount:   reviewCount,
			AverageRating: avgRating,
		})
	}

	// If no discover matches but has connections, default to connections tab
	if len(matchCards) == 0 && len(connectionCards) > 0 && activeTab == "discover" {
		activeTab = "connections"
	}

	flashes := auth.GetFlash(w, r)
	var successMsg string
	for _, flash := range flashes {
		if flash.Type == auth.FlashSuccess {
			successMsg = flash.Message
			break
		}
	}

	component := matchespages.MatchesPage(
		matchCards,
		connectionCards,
		pendingCards,
		activeTab,
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

	// Check if connected (mutual match)
	connected, _ := h.repo.AreConnected(r.Context(), sessionData.UserID, matchedUserID)
	if connected {
		// Award 'first_match' badge to both users
		// The service will check if it's their first match and send notifications
		currentUserID := uuid.MustParse(sessionData.UserID)
		otherUserID := uuid.MustParse(matchedUserID)

		// Award badge to current user
		if err := h.badgeService.AwardBadge(r.Context(), currentUserID, "first_match"); err != nil {
			// Log error but don't fail the request
			log.Printf("Failed to award first_match badge to user %s: %v", sessionData.UserID, err)
		}

		// Award badge to matched user
		if err := h.badgeService.AwardBadge(r.Context(), otherUserID, "first_match"); err != nil {
			log.Printf("Failed to award first_match badge to user %s: %v", matchedUserID, err)
		}
	}

	// For HTMX requests, return a notification snippet or empty to remove card
	if r.Header.Get("HX-Request") == "true" {
		if connected {
			// Show a success notification for mutual match
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "🎉 It's a match! You can now message each other.", "type": "success"}}`)
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	if connected {
		_ = auth.SetFlash(w, r, "🎉 It's a match! You can now message this user.", auth.FlashSuccess)
	} else {
		_ = auth.SetFlash(w, r, "Interest sent! You'll be connected when they accept too.", auth.FlashSuccess)
	}
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
