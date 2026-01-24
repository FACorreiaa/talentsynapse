package session

import (
	"time"
)

// Session represents a skill exchange session
type Session struct {
	ID              string    `json:"id"`
	InitiatorID     string    `json:"initiator_id"`
	PartnerID       string    `json:"partner_id"`
	InitiatorName   string    `json:"initiator_name"`   // Joined
	PartnerName     string    `json:"partner_name"`     // Joined
	InitiatorAvatar string    `json:"initiator_avatar"` // Joined
	PartnerAvatar   string    `json:"partner_avatar"`   // Joined
	Status          string    `json:"status"`
	ScheduledStart  time.Time `json:"scheduled_start"`
	IsReviewed      bool      `json:"is_reviewed"` // Computed field
}
