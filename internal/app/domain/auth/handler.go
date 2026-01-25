package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/user"
	"github.com/FACorreiaa/talentsynapse/internal/app/views/pages"
	authpages "github.com/FACorreiaa/talentsynapse/internal/app/views/pages/auth"
)

// Handler handles authentication HTTP requests
type Handler struct {
	authService ServiceInterface
}

// NewHandler creates a new auth handler
func NewHandler(authService ServiceInterface) *Handler {
	return &Handler{authService: authService}
}

// ShowLogin renders the login page
func (h *Handler) ShowLogin(w http.ResponseWriter, r *http.Request) {
	flashes := GetFlash(w, r)

	var errorMsg, successMsg string
	for _, flash := range flashes {
		switch flash.Type {
		case FlashError:
			errorMsg = flash.Message
		case FlashSuccess:
			successMsg = flash.Message
		}
	}

	component := authpages.Login(errorMsg, successMsg)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleLogin processes the login form
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		_ = SetFlash(w, r, "Invalid form data", FlashError)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	input := user.LoginInput{
		Email:    strings.TrimSpace(r.FormValue("email")),
		Password: r.FormValue("password"),
	}

	if input.Email == "" || input.Password == "" {
		_ = SetFlash(w, r, "Email and password are required", FlashError)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	u, err := h.authService.Login(r.Context(), input)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			_ = SetFlash(w, r, "Invalid email or password", FlashError)
		} else {
			_ = SetFlash(w, r, "An error occurred. Please try again.", FlashError)
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Check if MFA is enabled for this user
	if isMFAEnabled() {
		// Try to cast to concrete Service type to access mfaService
		if authSvc, ok := h.authService.(*Service); ok && authSvc.mfaService != nil {
			mfaEnabled, err := authSvc.mfaService.IsMFAEnabled(r.Context(), u.ID)
			if err == nil && mfaEnabled {
				// Store pending login in session
				session, _ := store.Get(r, SessionName)
				session.Values["pending_mfa_user_id"] = u.ID
				session.Values["pending_mfa_email"] = u.Email
				_ = session.Save(r, w)

				// Redirect to MFA verification
				http.Redirect(w, r, "/mfa/verify", http.StatusSeeOther)
				return
			}
		}
	}

	// No MFA or not enabled - complete login normally
	if err := CreateSession(w, r, u); err != nil {
		_ = SetFlash(w, r, "Failed to create session", FlashError)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	redirect := GetRedirectAfterLogin(w, r)
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

// ShowRegister renders the registration page
func (h *Handler) ShowRegister(w http.ResponseWriter, r *http.Request) {
	flashes := GetFlash(w, r)

	var errorMsg string
	for _, flash := range flashes {
		if flash.Type == FlashError {
			errorMsg = flash.Message
			break
		}
	}

	component := authpages.Register(errorMsg)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleRegister processes the registration form
func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		_ = SetFlash(w, r, "Invalid form data", FlashError)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	input := user.RegisterInput{
		DisplayName:     strings.TrimSpace(r.FormValue("name")),
		Email:           strings.TrimSpace(r.FormValue("email")),
		Username:        strings.TrimSpace(r.FormValue("username")),
		Password:        r.FormValue("password"),
		ConfirmPassword: r.FormValue("confirm_password"),
	}

	if input.DisplayName == "" || input.Email == "" || input.Username == "" || input.Password == "" {
		_ = SetFlash(w, r, "All fields are required", FlashError)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	u, err := h.authService.Register(r.Context(), input)
	if err != nil {
		var msg string
		switch {
		case errors.Is(err, ErrInvalidEmail):
			msg = "Please enter a valid email address"
		case errors.Is(err, ErrWeakPassword):
			msg = "Password must be at least 8 characters"
		case errors.Is(err, ErrPasswordMismatch):
			msg = "Passwords do not match"
		case errors.Is(err, user.ErrEmailAlreadyExists):
			msg = "An account with this email already exists"
		case errors.Is(err, user.ErrUsernameExists):
			msg = "This username is already taken"
		default:
			msg = "Failed to create account. Please try again."
		}
		_ = SetFlash(w, r, msg, FlashError)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	if err := CreateSession(w, r, u); err != nil {
		_ = SetFlash(w, r, "Account created! Please check your email to verify your account.", FlashSuccess)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	_ = SetFlash(w, r, "Account created! Please check your email to verify your account.", FlashSuccess)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// HandleLogout destroys the session
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	_ = DestroySession(w, r)
	_ = SetFlash(w, r, "You have been logged out", FlashSuccess)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ShowForgotPassword renders the forgot password page
func (h *Handler) ShowForgotPassword(w http.ResponseWriter, r *http.Request) {
	flashes := GetFlash(w, r)

	var errorMsg, successMsg string
	for _, flash := range flashes {
		switch flash.Type {
		case FlashError:
			errorMsg = flash.Message
		case FlashSuccess:
			successMsg = flash.Message
		}
	}

	component := authpages.ForgotPassword(errorMsg, successMsg)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleForgotPassword processes the forgot password form
func (h *Handler) HandleForgotPassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		_ = SetFlash(w, r, "Invalid form data", FlashError)
		http.Redirect(w, r, "/forgot-password", http.StatusSeeOther)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))

	if email == "" || !ValidateEmail(email) {
		_ = SetFlash(w, r, "Please enter a valid email address", FlashError)
		http.Redirect(w, r, "/forgot-password", http.StatusSeeOther)
		return
	}

	_ = h.authService.RequestPasswordReset(r.Context(), email)

	_ = SetFlash(w, r, "If an account exists with this email, you will receive a password reset link shortly.", FlashSuccess)
	http.Redirect(w, r, "/forgot-password", http.StatusSeeOther)
}

// ShowHome renders the home page
func ShowHome(w http.ResponseWriter, r *http.Request) {
	sessionData := GetSessionData(r)
	component := pages.Index(sessionData.IsAuthenticated, sessionData.UserName)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleEmailVerification processes the email verification link
func (h *Handler) HandleEmailVerification(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	if token == "" {
		_ = SetFlash(w, r, "Invalid verification link", FlashError)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, err := h.authService.VerifyEmail(r.Context(), token)
	if err != nil {
		var msg string
		if errors.Is(err, ErrTokenExpired) {
			msg = "Verification link has expired. Please request a new one."
		} else if errors.Is(err, ErrTokenNotFound) {
			msg = "Invalid or already used verification link."
		} else {
			msg = "Failed to verify email. Please try again."
		}
		_ = SetFlash(w, r, msg, FlashError)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Successfully verified
	_ = SetFlash(w, r, "Email verified successfully! You can now log in.", FlashSuccess)

	// If user is currently logged in, update their session
	sessionData := GetSessionData(r)
	if sessionData.IsAuthenticated && sessionData.UserID == userID {
		_ = UpdateSessionEmailVerified(w, r, true)
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// HandleResendVerification resends the verification email
func (h *Handler) HandleResendVerification(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		_ = SetFlash(w, r, "Invalid form data", FlashError)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))

	if email == "" || !ValidateEmail(email) {
		_ = SetFlash(w, r, "Please enter a valid email address", FlashError)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Always return success to prevent email enumeration
	_ = h.authService.ResendVerificationEmail(r.Context(), email)

	_ = SetFlash(w, r, "If an unverified account exists with this email, a verification link has been sent.", FlashSuccess)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// ShowResetPassword shows the reset password page
func (h *Handler) ShowResetPassword(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	if token == "" {
		_ = SetFlash(w, r, "Invalid reset link", FlashError)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	flashes := GetFlash(w, r)
	var errorMsg string
	for _, flash := range flashes {
		if flash.Type == FlashError {
			errorMsg = flash.Message
			break
		}
	}

	component := authpages.ResetPassword(token, errorMsg)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleResetPassword processes the password reset form
func (h *Handler) HandleResetPassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		_ = SetFlash(w, r, "Invalid form data", FlashError)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	token := r.FormValue("token")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if token == "" {
		_ = SetFlash(w, r, "Invalid reset link", FlashError)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if password == "" || confirmPassword == "" {
		_ = SetFlash(w, r, "Both password fields are required", FlashError)
		http.Redirect(w, r, "/reset-password?token="+token, http.StatusSeeOther)
		return
	}

	if password != confirmPassword {
		_ = SetFlash(w, r, "Passwords do not match", FlashError)
		http.Redirect(w, r, "/reset-password?token="+token, http.StatusSeeOther)
		return
	}

	// Reset the password
	err := h.authService.ResetPassword(r.Context(), token, password)
	if err != nil {
		var msg string
		if errors.Is(err, ErrTokenExpired) {
			msg = "Reset link has expired. Please request a new one."
		} else if errors.Is(err, ErrTokenNotFound) {
			msg = "Invalid or already used reset link."
		} else if errors.Is(err, ErrWeakPassword) {
			msg = "Password must be at least 8 characters long."
		} else {
			msg = "Failed to reset password. Please try again."
		}
		_ = SetFlash(w, r, msg, FlashError)
		http.Redirect(w, r, "/reset-password?token="+token, http.StatusSeeOther)
		return
	}

	// Success
	_ = SetFlash(w, r, "Password reset successfully! You can now log in with your new password.", FlashSuccess)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
