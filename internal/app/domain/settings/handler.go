package settings

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/auth"
	settingspages "github.com/FACorreiaa/talentsynapse/internal/app/views/pages/settings"
)

// Handler handles settings page HTTP requests
type Handler struct{}

// NewHandler creates a new settings handler
func NewHandler() *Handler {
	return &Handler{}
}

// Show renders the settings page
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	component := settingspages.Settings(
		sessionData.UserName,
		sessionData.UserAvatar,
		sessionData.UserEmail,
	)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// getSettingsData creates the common settings data struct
func (h *Handler) getSettingsData(r *http.Request) settingspages.SettingsData {
	sessionData := auth.GetSessionData(r)
	return settingspages.SettingsData{
		UserName:   sessionData.UserName,
		UserAvatar: sessionData.UserAvatar,
		UserEmail:  sessionData.UserEmail,
	}
}

// TabAccount renders the account tab content (HTMX partial)
func (h *Handler) TabAccount(w http.ResponseWriter, r *http.Request) {
	data := h.getSettingsData(r)
	data.ActiveTab = "account"

	component := settingspages.AccountTab(data)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// TabNotifications renders the notifications tab content (HTMX partial)
func (h *Handler) TabNotifications(w http.ResponseWriter, r *http.Request) {
	data := h.getSettingsData(r)
	data.ActiveTab = "notifications"

	component := settingspages.NotificationsTab(data)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// TabPrivacy renders the privacy tab content (HTMX partial)
func (h *Handler) TabPrivacy(w http.ResponseWriter, r *http.Request) {
	data := h.getSettingsData(r)
	data.ActiveTab = "privacy"

	component := settingspages.PrivacyTab(data)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// TabSecurity renders the security tab content (HTMX partial)
func (h *Handler) TabSecurity(w http.ResponseWriter, r *http.Request) {
	data := h.getSettingsData(r)
	data.ActiveTab = "security"

	// TODO: Check if MFA is enabled for user
	mfaEnabled := false

	component := settingspages.SecurityTabContent(data, mfaEnabled)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// TabAppearance renders the appearance tab content (HTMX partial)
func (h *Handler) TabAppearance(w http.ResponseWriter, r *http.Request) {
	data := h.getSettingsData(r)
	data.ActiveTab = "appearance"

	component := settingspages.AppearanceTab(data)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// TabDanger renders the danger zone tab content (HTMX partial)
func (h *Handler) TabDanger(w http.ResponseWriter, r *http.Request) {
	data := h.getSettingsData(r)
	data.ActiveTab = "danger"

	component := settingspages.DangerTab(data)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Tab handles dynamic tab routing
func (h *Handler) Tab(w http.ResponseWriter, r *http.Request) {
	tabName := chi.URLParam(r, "tab")

	switch tabName {
	case "account":
		h.TabAccount(w, r)
	case "notifications":
		h.TabNotifications(w, r)
	case "privacy":
		h.TabPrivacy(w, r)
	case "security":
		h.TabSecurity(w, r)
	case "appearance":
		h.TabAppearance(w, r)
	case "danger":
		h.TabDanger(w, r)
	default:
		// Default to account tab
		h.TabAccount(w, r)
	}
}
