package review

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	RequesterID uuid.UUID  `json:"requester_id" db:"requester_id"`
	ReviewerID  *uuid.UUID `json:"reviewer_id" db:"reviewer_id"`
	SkillID     uuid.UUID  `json:"skill_id" db:"skill_id"` // For MVP, we might associate with a generic skill or handle nulled if schema allows (schema says NOT NULL)

	// Ratings
	RequesterRating  *int    `json:"requester_rating" db:"requester_rating"`
	RequesterComment *string `json:"requester_comment" db:"requester_comment"`
	ReviewerRating   *int    `json:"reviewer_rating" db:"reviewer_rating"`
	ReviewerComment  *string `json:"reviewer_comment" db:"reviewer_comment"`

	CreatedAt time.Time `json:"created_at" db:"requested_at"` // using requested_at as creation time
}

type UserWithReview struct {
	Rating         int
	Comment        string
	ReviewerName   string
	ReviewerAvatar string
	CreatedAt      time.Time
}
