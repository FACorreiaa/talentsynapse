package badges

import (
	"context"
	"log"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CheckMilestones evaluates user stats and awards badges if criteria met.
// This is a simplified version; normally would accept stats or be triggered by events.
func (s *Service) CheckMilestones(ctx context.Context, userID uuid.UUID) {
	// Example: Award 'early_adopter' to everyone for now (or check creation date)
	// Real implementation would inject UserStats repo to check "total_sessions" etc.

	// For MVP demonstration, let's just ensure they have the 'early_adopter' badge.
	if err := s.repo.AwardBadge(ctx, userID, "early_adopter"); err != nil {
		log.Printf("Error awarding early_adopter badge: %v", err)
	}

	// TODO: Add logic for 'first_match' when a match is accepted.
	// TODO: Add logic for 'top_teacher' when rating > 4.5.
}

func (s *Service) GetUserBadges(ctx context.Context, userID uuid.UUID) ([]*UserBadge, error) {
	return s.repo.GetUserBadges(ctx, userID)
}
