package profile

import (
	"net/http"

	"github.com/FACorreiaa/skillsphere/internal/app/auth"
	"github.com/FACorreiaa/skillsphere/internal/app/user"
	profilepages "github.com/FACorreiaa/skillsphere/internal/app/views/pages/profile"
)

// Handler handles profile HTTP requests
type Handler struct {
	userRepo *user.Repository
}

// NewHandler creates a new profile handler
func NewHandler(userRepo *user.Repository) *Handler {
	return &Handler{userRepo: userRepo}
}

// Show renders the user profile page
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get full user data
	userData, err := h.userRepo.GetByID(r.Context(), sessionData.UserID)
	if err != nil {
		http.Error(w, "Failed to load profile", http.StatusInternalServerError)
		return
	}

	// Build profile view model
	profile := profilepages.ProfileData{
		ID:          userData.ID,
		Username:    userData.Username,
		DisplayName: userData.DisplayName,
		Email:       userData.Email,
		AvatarURL:   getStringPtr(userData.AvatarURL),
		CreatedAt:   userData.CreatedAt,
	}

	flashes := auth.GetFlash(w, r)
	var successMsg string
	for _, flash := range flashes {
		if flash.Type == auth.FlashSuccess {
			successMsg = flash.Message
			break
		}
	}

	component := profilepages.Profile(profile, sessionData.UserName, sessionData.UserAvatar, successMsg)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Update handles profile updates via form POST
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	displayName := r.FormValue("display_name")
	bio := r.FormValue("bio")

	err := h.userRepo.UpdateProfile(r.Context(), sessionData.UserID, displayName, bio)
	if err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	_ = auth.SetFlash(w, r, "Profile updated successfully!", auth.FlashSuccess)
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func getStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
