package types

// SessionData represents the current user's session information
// This is used across the application for passing user context to handlers and templates
type SessionData struct {
	IsAuthenticated bool
	UserID          string
	UserName        string
	UserEmail       string
	UserAvatar      string
	Role            string
	EmailVerified   bool
}
