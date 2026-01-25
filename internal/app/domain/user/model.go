package user

import (
	"time"

	"github.com/FACorreiaa/talentsynapse/internal/app/types"
)

// User represents a user entity in the system
type User struct {
	ID              string
	Email           string
	Username        string
	HashedPassword  string
	DisplayName     string
	AvatarURL       *string
	Bio             *string
	Role            string
	Tier            string // From user_stats
	IsActive        bool
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastLoginAt     *time.Time
	SocialLinks     SocialLinks
}

type SocialLinks struct {
	GitHub   string `json:"github,omitempty"`
	LinkedIn string `json:"linkedin,omitempty"`
	Twitter  string `json:"twitter,omitempty"`
	Website  string `json:"website,omitempty"`
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

// SessionData is an alias to the shared types.SessionData
// Kept for backward compatibility
type SessionData = types.SessionData

// PublicProfile represents a user's public-facing profile
type PublicProfile struct {
	ID          string
	Username    string
	DisplayName string
	AvatarURL   string
	Bio         string
	CreatedAt   time.Time
}
