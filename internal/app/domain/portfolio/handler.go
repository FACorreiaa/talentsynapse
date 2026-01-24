package portfolio

import (
	"net/http"

	"github.com/FACorreiaa/skillsphere/internal/app/domain/auth"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// Create handles creating a portfolio item
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	desc := r.FormValue("description")
	link := r.FormValue("link_url")

	// Image handling omitted for MVP (placeholder or external URL input)
	img := r.FormValue("image_url")

	_, err := h.repo.CreateItem(r.Context(), sessionData.UserID, title, desc, link, img)
	if err != nil {
		_ = auth.SetFlash(w, r, "Failed to add project", auth.FlashError)
	} else {
		_ = auth.SetFlash(w, r, "Project added successfully", auth.FlashSuccess)
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

// Delete handles removing a portfolio item
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	itemID := chi.URLParam(r, "id")
	if itemID == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.DeleteItem(r.Context(), itemID, sessionData.UserID); err != nil {
		_ = auth.SetFlash(w, r, "Failed to delete project", auth.FlashError)
	} else {
		_ = auth.SetFlash(w, r, "Project deleted", auth.FlashSuccess)
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
