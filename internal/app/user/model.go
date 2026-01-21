package user

import (
	"time"
)

// User represents a user entity in the system
type User struct {
	ID              string
	Email           string
	Username        string
	HashedPassword  string
	DisplayName     string
	AvatarURL       *string
	Role            string
	IsActive        bool
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastLoginAt     *time.Time
}

// CreateUserInput represents input for creating a new user
type CreateUserInput struct {
	Email       string
	Username    string
	Password    string
	DisplayName string
}

// UpdateUserInput represents input for updating a user
type UpdateUserInput struct {
	DisplayName *string
	AvatarURL   *string
	Bio         *string
}

// LoginInput represents input for login
type LoginInput struct {
	Email    string
	Password string
}

// RegisterInput represents input for registration
type RegisterInput struct {
	Email           string
	Username        string
	Password        string
	ConfirmPassword string
	DisplayName     string
}

// SessionData represents user session data for templates
type SessionData struct {
	IsAuthenticated bool
	UserID          string
	UserName        string
	UserEmail       string
	UserAvatar      string
}

// PublicProfile represents a user's public-facing profile
type PublicProfile struct {
	ID          string
	Username    string
	DisplayName string
	AvatarURL   string
	Bio         string
	CreatedAt   time.Time
}
