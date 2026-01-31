package review

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/auth"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/badges"
	"github.com/google/uuid"
)

type Handler struct {
	repo       *Repository
	badgesRepo *badges.Repository
}

func NewHandler(repo *Repository, badgesRepo *badges.Repository) *Handler {
	return &Handler{
		repo:       repo,
		badgesRepo: badgesRepo,
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

		// Gamification Check: Top Teacher
		// 5+ reviews, 4.5+ average
		total, avg, err := h.repo.GetStats(r.Context(), toUserID)
		if err == nil && total >= 5 && avg >= 4.5 {
			_ = h.badgesRepo.AwardBadge(r.Context(), toUserID, "top_teacher")
		}
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}
