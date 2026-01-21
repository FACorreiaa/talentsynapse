package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/sessions"

	"github.com/FACorreiaa/skillsphere/internal/app/user"
)

const (
	SessionName   = "skillsphere_session"
	UserIDKey     = "user_id"
	UserNameKey   = "user_name"
	UserEmailKey  = "user_email"
	UserAvatarKey = "user_avatar"
	FlashKey      = "flash"
	FlashTypeKey  = "flash_type"
)

// FlashType represents the type of flash message
type FlashType string

const (
	FlashSuccess FlashType = "success"
	FlashError   FlashType = "error"
	FlashInfo    FlashType = "info"
	FlashWarning FlashType = "warning"
)

// FlashMessage represents a flash message with its type
type FlashMessage struct {
	Message string
	Type    FlashType
}

var store *sessions.CookieStore

// InitStore initializes the session store
func InitStore() {
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		secret = generateRandomSecret()
	}

	store = sessions.NewCookieStore([]byte(secret))

	maxAge := 86400
	if envMaxAge := os.Getenv("SESSION_MAX_AGE"); envMaxAge != "" {
		if parsed, err := strconv.Atoi(envMaxAge); err == nil {
			maxAge = parsed
		}
	}

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   os.Getenv("GO_ENV") != "development",
		SameSite: http.SameSiteLaxMode,
	}
}

// GetStore returns the session store
func GetStore() *sessions.CookieStore {
	if store == nil {
		InitStore()
	}
	return store
}

// GetSession retrieves the session from the request
func GetSession(r *http.Request) (*sessions.Session, error) {
	return GetStore().Get(r, SessionName)
}

// CreateSession creates a new authenticated session
func CreateSession(w http.ResponseWriter, r *http.Request, u *user.User) error {
	session, err := GetSession(r)
	if err != nil {
		return err
	}

	session.Values[UserIDKey] = u.ID
	session.Values[UserNameKey] = u.DisplayName
	session.Values[UserEmailKey] = u.Email
	if u.AvatarURL != nil {
		session.Values[UserAvatarKey] = *u.AvatarURL
	} else {
		session.Values[UserAvatarKey] = ""
	}

	return session.Save(r, w)
}

// DestroySession clears the session
func DestroySession(w http.ResponseWriter, r *http.Request) error {
	session, err := GetSession(r)
	if err != nil {
		return err
	}

	session.Values = make(map[interface{}]interface{})
	session.Options.MaxAge = -1

	return session.Save(r, w)
}

// IsAuthenticated checks if user is authenticated
func IsAuthenticated(r *http.Request) bool {
	session, err := GetSession(r)
	if err != nil {
		return false
	}

	userID, ok := session.Values[UserIDKey].(string)
	return ok && userID != ""
}

// GetUserID returns the user ID from session
func GetUserID(r *http.Request) string {
	session, err := GetSession(r)
	if err != nil {
		return ""
	}

	if userID, ok := session.Values[UserIDKey].(string); ok {
		return userID
	}
	return ""
}

// GetSessionData returns all session data for templates
func GetSessionData(r *http.Request) user.SessionData {
	session, err := GetSession(r)
	if err != nil {
		return user.SessionData{}
	}

	data := user.SessionData{
		IsAuthenticated: IsAuthenticated(r),
	}

	if userID, ok := session.Values[UserIDKey].(string); ok {
		data.UserID = userID
	}
	if userName, ok := session.Values[UserNameKey].(string); ok {
		data.UserName = userName
	}
	if userEmail, ok := session.Values[UserEmailKey].(string); ok {
		data.UserEmail = userEmail
	}
	if userAvatar, ok := session.Values[UserAvatarKey].(string); ok {
		data.UserAvatar = userAvatar
	}

	return data
}

// SetFlash sets a flash message
func SetFlash(w http.ResponseWriter, r *http.Request, message string, flashType FlashType) error {
	session, err := GetSession(r)
	if err != nil {
		return err
	}

	session.AddFlash(message, FlashKey)
	session.AddFlash(string(flashType), FlashTypeKey)

	return session.Save(r, w)
}

// GetFlash retrieves and clears flash messages
func GetFlash(w http.ResponseWriter, r *http.Request) []FlashMessage {
	session, err := GetSession(r)
	if err != nil {
		return nil
	}

	messages := session.Flashes(FlashKey)
	types := session.Flashes(FlashTypeKey)

	if len(messages) == 0 {
		return nil
	}

	session.Save(r, w)

	result := make([]FlashMessage, 0, len(messages))
	for i, msg := range messages {
		flashType := FlashInfo
		if i < len(types) {
			if t, ok := types[i].(string); ok {
				flashType = FlashType(t)
			}
		}
		if m, ok := msg.(string); ok {
			result = append(result, FlashMessage{Message: m, Type: flashType})
		}
	}

	return result
}

// GetRedirectAfterLogin retrieves and clears redirect URL
func GetRedirectAfterLogin(w http.ResponseWriter, r *http.Request) string {
	session, err := GetSession(r)
	if err != nil {
		return "/dashboard"
	}

	redirect, ok := session.Values["redirect_after_login"].(string)
	if !ok || redirect == "" {
		return "/dashboard"
	}

	delete(session.Values, "redirect_after_login")
	session.Save(r, w)

	return redirect
}

func generateRandomSecret() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.StdEncoding.EncodeToString(bytes)
}
