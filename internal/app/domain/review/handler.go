package review

import (
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
		_ = auth.SetFlash(w, r, "Invalid rating", auth.FlashError)
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}

	comment := r.FormValue("comment")
	sessionID := r.FormValue("session_id")

	if err := h.repo.CreateSessionReview(r.Context(), fromUserID.String(), toUserID.String(), sessionID, rating, comment); err != nil {
		_ = auth.SetFlash(w, r, "Failed to submit review", auth.FlashError)
	} else {
		_ = auth.SetFlash(w, r, "Review submitted matched!", auth.FlashSuccess)

		// Gamification Check: Top Teacher
		// 5+ reviews, 4.5+ average
		total, avg, err := h.repo.GetStats(r.Context(), toUserID)
		if err == nil && total >= 5 && avg >= 4.5 {
			_ = h.badgesRepo.AwardBadge(r.Context(), toUserID, "top_teacher")
		}
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}
