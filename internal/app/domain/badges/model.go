package badges

import (
	"time"

	"github.com/google/uuid"
)

type Badge struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	IconURL     string    `json:"icon_url" db:"icon_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type UserBadge struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	BadgeID   uuid.UUID `json:"badge_id" db:"badge_id"`
	AwardedAt time.Time `json:"awarded_at" db:"awarded_at"`

	// Joined fields
	BadgeCode string `json:"badge_code" db:"badge_code"`
	BadgeName string `json:"badge_name" db:"badge_name"`
	BadgeIcon string `json:"badge_icon" db:"badge_icon"`
}
