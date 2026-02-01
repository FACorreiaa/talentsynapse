package review

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/auth"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/badges"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/points"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/push"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/user"
	"github.com/google/uuid"
)

// NotificationHub interface for sending real-time notifications
type NotificationHub interface {
	BroadcastToUser(userID string, message []byte)
}

type Handler struct {
	repo          *Repository
	badgesRepo    *badges.Repository
	pointsService *points.Service
	userRepo      *user.Repository
	hub           NotificationHub
	pushService   *push.Service
}

func NewHandler(repo *Repository, badgesRepo *badges.Repository, pointsService *points.Service, userRepo *user.Repository, hub NotificationHub, pushService *push.Service) *Handler {
	return &Handler{
		repo:          repo,
		badgesRepo:    badgesRepo,
		pointsService: pointsService,
		userRepo:      userRepo,
		hub:           hub,
		pushService:   pushService,
	}
}

func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	session := auth.GetSessionData(r)
	if session.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	fromUserID := uuid.MustParse(session.UserID)

	toUserIDStr := r.FormValue("target_user_id")
	toUserID, err := uuid.Parse(toUserIDStr)
	if err != nil {
		_ = auth.SetFlash(w, r, "Invalid user", auth.FlashError)
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}

	ratingStr := r.FormValue("rating")
	rating, err := strconv.Atoi(ratingStr)
	if err != nil || rating < 1 || rating > 5 {
		_ = auth.SetFlash(w, r, "Invalid rating (must be 1-5)", auth.FlashError)
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}

	comment := r.FormValue("comment")
	sessionID := r.FormValue("session_id")

	// Validate session_id is provided
	if sessionID == "" {
		_ = auth.SetFlash(w, r, "Session ID is required", auth.FlashError)
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}

	// Validate the session_id is a valid UUID
	if _, err := uuid.Parse(sessionID); err != nil {
		_ = auth.SetFlash(w, r, "Invalid session ID", auth.FlashError)
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}

	// Validate review submission (checks session status, duplicates, participation)
	if err := h.repo.ValidateReviewSubmission(r.Context(), fromUserID.String(), toUserID.String(), sessionID); err != nil {
		// Map specific errors to user-friendly messages
		var flashMsg string
		switch {
		case errors.Is(err, ErrDuplicateReview):
			flashMsg = "You have already submitted a review for this session"
		case errors.Is(err, ErrSessionNotCompleted):
			flashMsg = "You can only submit a review after the session is completed"
		case errors.Is(err, ErrSessionNotFound):
			flashMsg = "Session not found"
		case errors.Is(err, ErrNotSessionParticipant):
			flashMsg = "You can only review sessions you participated in"
		case errors.Is(err, ErrCannotReviewSelf):
			flashMsg = "You cannot review yourself"
		default:
			flashMsg = "Failed to validate review submission"
		}
		_ = auth.SetFlash(w, r, flashMsg, auth.FlashError)
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}

	// All validations passed, create the review
	if err := h.repo.CreateSessionReview(r.Context(), fromUserID.String(), toUserID.String(), sessionID, rating, comment); err != nil {
		_ = auth.SetFlash(w, r, "Failed to submit review", auth.FlashError)
	} else {
		_ = auth.SetFlash(w, r, "Review submitted successfully!", auth.FlashSuccess)

		// Send notification to the reviewed user
		h.notifyReviewReceived(r.Context(), toUserID, fromUserID, rating)

		// Award points to the reviewed user based on rating
		if h.pointsService != nil {
			upgrade, err := h.pointsService.AwardReviewPoints(r.Context(), toUserID, rating)
			if err != nil {
				log.Printf("Failed to award review points: %v", err)
			} else if upgrade != nil {
				log.Printf("User %s upgraded to %s tier after review!", toUserID, upgrade.NewTier)
			}
		}

		// Gamification Check: Top Teacher
		// 5+ reviews, 4.5+ average
		total, avg, err := h.repo.GetStats(r.Context(), toUserID)
		if err == nil && total >= 5 && avg >= 4.5 {
			_ = h.badgesRepo.AwardBadge(r.Context(), toUserID, "top_teacher")
		}
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

// notifyReviewReceived sends a real-time notification to the user who received a review
func (h *Handler) notifyReviewReceived(ctx context.Context, reviewedUserID, reviewerUserID uuid.UUID, rating int) {
	if h.hub == nil || h.userRepo == nil {
		log.Printf("Review notification skipped: hub or userRepo not configured")
		return
	}

	// Get reviewer information
	reviewer, err := h.userRepo.GetByID(ctx, reviewerUserID.String())
	if err != nil {
		log.Printf("Failed to get reviewer info for notification: %v", err)
		return
	}

	// Generate star icons based on rating
	stars := ""
	for i := 0; i < rating; i++ {
		stars += "⭐"
	}

	// Create notification HTML
	notification := fmt.Sprintf(`
		<div id="notifications" hx-swap-oob="beforeend">
			<div class="alert shadow-lg mb-2 bg-gradient-to-r from-blue-500 to-purple-500 text-white animate-fade-in" role="alert">
				<div class="flex items-center gap-3">
					<span class="text-3xl">📝</span>
					<div>
						<h3 class="font-bold text-lg">New Review Received!</h3>
						<div class="text-sm"><strong>%s</strong> left you a review</div>
						<div class="text-sm mt-1">Rating: %s</div>
						<a href="/users/%s" class="text-xs underline hover:text-blue-200 mt-1 inline-block">View your profile</a>
					</div>
				</div>
				<button class="btn btn-sm btn-ghost text-white" onclick="this.parentElement.remove()">✕</button>
			</div>
		</div>
	`, reviewer.DisplayName, stars, reviewedUserID.String())

	h.hub.BroadcastToUser(reviewedUserID.String(), []byte(notification))
	log.Printf("📝 Review notification sent: %s reviewed %s with %d stars", reviewer.DisplayName, reviewedUserID, rating)

	// Send push notification
	if h.pushService != nil {
		if err := h.pushService.SendNewReviewNotification(ctx, reviewedUserID, reviewer.DisplayName, rating); err != nil {
			log.Printf("Failed to send review push notification to user %s: %v", reviewedUserID, err)
		}
	}
}
