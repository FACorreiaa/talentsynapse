package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/FACorreiaa/skillsphere-pwa/internal/app/user"
	"github.com/FACorreiaa/skillsphere-pwa/views/pages"
	authpages "github.com/FACorreiaa/skillsphere-pwa/views/pages/auth"
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
		SetFlash(w, r, "Invalid form data", FlashError)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	input := user.LoginInput{
		Email:    strings.TrimSpace(r.FormValue("email")),
		Password: r.FormValue("password"),
	}

	if input.Email == "" || input.Password == "" {
		SetFlash(w, r, "Email and password are required", FlashError)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	u, err := h.authService.Login(r.Context(), input)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			SetFlash(w, r, "Invalid email or password", FlashError)
		} else {
			SetFlash(w, r, "An error occurred. Please try again.", FlashError)
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := CreateSession(w, r, u); err != nil {
		SetFlash(w, r, "Failed to create session", FlashError)
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
		SetFlash(w, r, "Invalid form data", FlashError)
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
		SetFlash(w, r, "All fields are required", FlashError)
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
		SetFlash(w, r, msg, FlashError)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	if err := CreateSession(w, r, u); err != nil {
		SetFlash(w, r, "Account created! Please log in.", FlashSuccess)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// HandleLogout destroys the session
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	_ = DestroySession(w, r)
	SetFlash(w, r, "You have been logged out", FlashSuccess)
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
		SetFlash(w, r, "Invalid form data", FlashError)
		http.Redirect(w, r, "/forgot-password", http.StatusSeeOther)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))

	if email == "" || !ValidateEmail(email) {
		SetFlash(w, r, "Please enter a valid email address", FlashError)
		http.Redirect(w, r, "/forgot-password", http.StatusSeeOther)
		return
	}

	_ = h.authService.RequestPasswordReset(r.Context(), email)

	SetFlash(w, r, "If an account exists with this email, you will receive a password reset link shortly.", FlashSuccess)
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
