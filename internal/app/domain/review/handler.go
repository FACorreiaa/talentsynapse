package review

import (
	"net/http"
	"strconv"

	"github.com/FACorreiaa/skillsphere/internal/app/domain/auth"
	"github.com/google/uuid"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
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

	// Create review
	if err := h.repo.CreateSimpleReview(r.Context(), fromUserID, toUserID, rating, comment); err != nil {
		_ = auth.SetFlash(w, r, "Failed to submit review", auth.FlashError)
	} else {
		_ = auth.SetFlash(w, r, "Review submitted matched!", auth.FlashSuccess)
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}
