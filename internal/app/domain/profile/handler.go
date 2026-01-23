package profile

import (
	"net/http"

	"github.com/FACorreiaa/skillsphere/internal/app/domain/auth"
	"github.com/FACorreiaa/skillsphere/internal/app/domain/user"
	profilepages "github.com/FACorreiaa/skillsphere/internal/app/views/pages/profile"
	"github.com/go-chi/chi/v5"
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

// ShowPublic renders another user's public profile
func (h *Handler) ShowPublic(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get profile user ID from URL
	profileUserID := chi.URLParam(r, "id")
	if profileUserID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	// Redirect to own profile if viewing self
	if profileUserID == sessionData.UserID {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	// Get target user data
	targetUser, err := h.userRepo.GetByID(r.Context(), profileUserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get user's skills
	offeredSkills, wantedSkills, err := h.userRepo.GetUserSkillNames(r.Context(), profileUserID)
	if err != nil {
		offeredSkills = []string{}
		wantedSkills = []string{}
	}

	// Check if users are connected (mutual match)
	isConnected, _ := h.userRepo.AreConnected(r.Context(), sessionData.UserID, profileUserID)

	profile := profilepages.PublicProfileData{
		ID:            targetUser.ID,
		Username:      targetUser.Username,
		DisplayName:   targetUser.DisplayName,
		AvatarURL:     getStringPtr(targetUser.AvatarURL),
		Bio:           getStringPtr(targetUser.Bio),
		OfferedSkills: offeredSkills,
		WantedSkills:  wantedSkills,
		IsConnected:   isConnected,
	}

	component := profilepages.PublicProfile(profile, sessionData.UserName, sessionData.UserAvatar)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
