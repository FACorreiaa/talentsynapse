package auth

import (
	"fmt"
	"net/http"
	"os"

	"github.com/FACorreiaa/talentsynapse/internal/app/views/pages/auth"
	"github.com/FACorreiaa/talentsynapse/internal/app/views/pages/settings"
)

// HandleMFASetup shows the MFA setup page
func (h *Handler) HandleMFASetup(w http.ResponseWriter, r *http.Request) {
	if !isMFAEnabled() {
		http.Redirect(w, r, "/settings/security", http.StatusSeeOther)
		return
	}

	sessionData := GetSessionData(r)
	if sessionData.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	authSvc, ok := h.authService.(*Service)
	if !ok {
		http.Error(w, "Service not available", http.StatusInternalServerError)
		return
	}

	setup, err := authSvc.mfaService.GenerateSetup(r.Context(), sessionData.UserID, sessionData.UserEmail, sessionData.UserName)
	if err != nil {
		http.Error(w, "Failed to generate MFA setup", http.StatusInternalServerError)
		return
	}

	data := auth.MFASetupData{
		QRCodeURL:   setup.QRCodeURL,
		Secret:      setup.Secret,
		BackupCodes: setup.BackupCodes,
	}
	if err := auth.MFASetup(sessionData, data).Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// HandleMFAVerifySetup verifies the setup code and enables MFA
func (h *Handler) HandleMFAVerifySetup(w http.ResponseWriter, r *http.Request) {
	if !isMFAEnabled() {
		http.Redirect(w, r, "/settings/security", http.StatusSeeOther)
		return
	}

	sessionData := GetSessionData(r)
	if sessionData.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code := r.FormValue("code")
	if code == "" {
		_ = SetFlash(w, r, "Verification code is required", FlashError)
		http.Redirect(w, r, "/mfa/setup", http.StatusSeeOther)
		return
	}

	authSvc, ok := h.authService.(*Service)
	if !ok {
		http.Error(w, "Service not available", http.StatusInternalServerError)
		return
	}

	err := authSvc.mfaService.EnableMFA(r.Context(), sessionData.UserID, code)
	if err != nil {
		_ = SetFlash(w, r, "Invalid verification code. Please try again.", FlashError)
		http.Redirect(w, r, "/mfa/setup", http.StatusSeeOther)
		return
	}

	_ = SetFlash(w, r, "Two-factor authentication has been enabled successfully!", FlashSuccess)
	http.Redirect(w, r, "/settings/security", http.StatusSeeOther)
}

// HandleMFAVerifyPage shows the MFA verification page during login
func (h *Handler) HandleMFAVerifyPage(w http.ResponseWriter, r *http.Request) {
	if !isMFAEnabled() {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	session, err := store.Get(r, SessionName)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	email, ok := session.Values["pending_mfa_email"].(string)
	if !ok || email == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := auth.MFAVerify(email).Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// HandleMFAVerifyCode verifies MFA code during login
func (h *Handler) HandleMFAVerifyCode(w http.ResponseWriter, r *http.Request) {
	if !isMFAEnabled() {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := store.Get(r, SessionName)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, ok := session.Values["pending_mfa_user_id"].(string)
	if !ok || userID == "" {
		_ = SetFlash(w, r, "Session expired. Please log in again.", FlashError)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	code := r.FormValue("code")
	if code == "" {
		_ = SetFlash(w, r, "Verification code is required", FlashError)
		http.Redirect(w, r, "/mfa/verify", http.StatusSeeOther)
		return
	}

	authSvc, ok := h.authService.(*Service)
	if !ok {
		http.Error(w, "Service not available", http.StatusInternalServerError)
		return
	}

	valid, err := authSvc.mfaService.VerifyTOTP(r.Context(), userID, code)
	if err != nil || !valid {
		_ = SetFlash(w, r, "Invalid verification code. Please try again.", FlashError)
		http.Redirect(w, r, "/mfa/verify", http.StatusSeeOther)
		return
	}

	h.completeMFALogin(w, r, userID)
}

// HandleMFAVerifyBackupCode verifies backup code during login
func (h *Handler) HandleMFAVerifyBackupCode(w http.ResponseWriter, r *http.Request) {
	if !isMFAEnabled() {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := store.Get(r, SessionName)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, ok := session.Values["pending_mfa_user_id"].(string)
	if !ok || userID == "" {
		_ = SetFlash(w, r, "Session expired. Please log in again.", FlashError)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	backupCode := r.FormValue("backup_code")
	if backupCode == "" {
		_ = SetFlash(w, r, "Backup code is required", FlashError)
		http.Redirect(w, r, "/mfa/verify", http.StatusSeeOther)
		return
	}

	authSvc, ok := h.authService.(*Service)
	if !ok {
		http.Error(w, "Service not available", http.StatusInternalServerError)
		return
	}

	valid, err := authSvc.mfaService.VerifyBackupCode(r.Context(), userID, backupCode)
	if err != nil || !valid {
		_ = SetFlash(w, r, "Invalid backup code. Please try again.", FlashError)
		http.Redirect(w, r, "/mfa/verify", http.StatusSeeOther)
		return
	}

	h.completeMFALogin(w, r, userID)
}

// completeMFALogin completes the login process after MFA verification
func (h *Handler) completeMFALogin(w http.ResponseWriter, r *http.Request, userID string) {
	authSvc, ok := h.authService.(*Service)
	if !ok {
		http.Error(w, "Service not available", http.StatusInternalServerError)
		return
	}

	user, err := authSvc.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		_ = SetFlash(w, r, "Failed to complete login", FlashError)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	session, _ := store.Get(r, SessionName)

	delete(session.Values, "pending_mfa_user_id")
	delete(session.Values, "pending_mfa_email")

	session.Values[UserIDKey] = user.ID
	session.Values[UserNameKey] = user.DisplayName
	session.Values[UserEmailKey] = user.Email
	session.Values[UserAvatarKey] = user.AvatarURL
	session.Values[EmailVerifiedKey] = user.EmailVerifiedAt != nil

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	_ = SetFlash(w, r, "Successfully logged in!", FlashSuccess)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// HandleMFADisable disables MFA for the current user
func (h *Handler) HandleMFADisable(w http.ResponseWriter, r *http.Request) {
	if !isMFAEnabled() {
		http.Redirect(w, r, "/settings/security", http.StatusSeeOther)
		return
	}

	sessionData := GetSessionData(r)
	if sessionData.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authSvc, ok := h.authService.(*Service)
	if !ok {
		http.Error(w, "Service not available", http.StatusInternalServerError)
		return
	}

	err := authSvc.mfaService.DisableMFA(r.Context(), sessionData.UserID)
	if err != nil {
		_ = SetFlash(w, r, "Failed to disable MFA", FlashError)
		http.Redirect(w, r, "/settings/security", http.StatusSeeOther)
		return
	}

	_ = SetFlash(w, r, "Two-factor authentication has been disabled", FlashSuccess)
	http.Redirect(w, r, "/settings/security", http.StatusSeeOther)
}

// HandleMFARegenerateBackupCodes generates new backup codes
func (h *Handler) HandleMFARegenerateBackupCodes(w http.ResponseWriter, r *http.Request) {
	if !isMFAEnabled() {
		http.Redirect(w, r, "/settings/security", http.StatusSeeOther)
		return
	}

	sessionData := GetSessionData(r)
	if sessionData.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authSvc, ok := h.authService.(*Service)
	if !ok {
		http.Error(w, "Service not available", http.StatusInternalServerError)
		return
	}

	codes, err := authSvc.mfaService.RegenerateBackupCodes(r.Context(), sessionData.UserID)
	if err != nil {
		_ = SetFlash(w, r, "Failed to regenerate backup codes", FlashError)
		http.Redirect(w, r, "/settings/security", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if _, err := fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
	<title>New Backup Codes | SkillSphere</title>
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gradient-to-br from-purple-50 via-white to-green-50 min-h-screen flex items-center justify-center p-4">
	<div class="max-w-2xl w-full bg-white rounded-lg shadow-lg p-8">
		<h1 class="text-2xl font-bold text-gray-900 mb-4">New Backup Codes Generated</h1>
		<div class="bg-amber-50 border-l-4 border-amber-400 p-4 mb-6">
			<p class="text-sm text-amber-800 font-medium">
				⚠️ Save these codes now! You won't be able to see them again.
			</p>
		</div>
		<div class="bg-gray-50 p-4 rounded-lg mb-6">
			<div class="grid grid-cols-2 gap-3">
	`); err != nil {
		return
	}

	for _, code := range codes {
		if _, err := fmt.Fprintf(w, `
				<div class="p-3 bg-white border border-gray-200 rounded-md">
					<code class="font-mono text-sm text-gray-900">%s</code>
				</div>
		`, code); err != nil {
			return
		}
	}

	_, _ = fmt.Fprintf(w, `
			</div>
		</div>
		<div class="flex gap-3">
			<button onclick="copyBackupCodes()" class="px-4 py-2 bg-purple-600 text-white rounded-md hover:bg-purple-700">
				Copy All
			</button>
			<a href="/settings/security" class="px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300">
				Done
			</a>
		</div>
	</div>
	<script>
		function copyBackupCodes() {
			const codes = Array.from(document.querySelectorAll('code'))
				.map(el => el.textContent)
				.join('\n');
			navigator.clipboard.writeText(codes).then(() => {
				alert('Backup codes copied to clipboard!');
			});
		}
	</script>
</body>
</html>
	`)
}

// HandleSecuritySettings shows the security settings page
func (h *Handler) HandleSecuritySettings(w http.ResponseWriter, r *http.Request) {
	sessionData := GetSessionData(r)
	if sessionData.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var mfaEnabled bool
	var backupCodesCount int

	if isMFAEnabled() {
		authSvc, ok := h.authService.(*Service)
		if ok {
			mfaEnabled, _ = authSvc.mfaService.IsMFAEnabled(r.Context(), sessionData.UserID)
			if mfaEnabled {
				backupCodesCount, _ = authSvc.mfaService.GetBackupCodesCount(r.Context(), sessionData.UserID)
			}
		}
	}

	data := settings.SecuritySettingsData{
		MFAEnabled:       mfaEnabled,
		BackupCodesCount: backupCodesCount,
	}

	if err := settings.SecuritySettings(sessionData, data).Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// isMFAEnabled checks if MFA is enabled based on environment
func isMFAEnabled() bool {
	env := os.Getenv("GO_ENV")
	mfaEnv := os.Getenv("MFA_ENABLED")
	return env == "production" || mfaEnv == "true"
}
