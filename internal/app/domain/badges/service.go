package badges

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
)

// NotificationHub interface for sending notifications
type NotificationHub interface {
	BroadcastToUser(userID string, message []byte)
}

type Service struct {
	repo *Repository
	hub  NotificationHub
}

func NewService(repo *Repository, hub NotificationHub) *Service {
	return &Service{
		repo: repo,
		hub:  hub,
	}
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

// AwardBadge awards a badge to a user and sends notification if it's new
func (s *Service) AwardBadge(ctx context.Context, userID uuid.UUID, badgeCode string) error {
	// Check if user already has this badge
	hasBadge, err := s.repo.HasBadge(ctx, userID, badgeCode)
	if err != nil {
		return err
	}

	// If already has badge, don't award again
	if hasBadge {
		return nil
	}

	// Award the badge
	if err := s.repo.AwardBadge(ctx, userID, badgeCode); err != nil {
		return err
	}

	// Get badge details for notification
	badge, err := s.repo.GetBadgeByCode(ctx, badgeCode)
	if err != nil {
		log.Printf("Failed to get badge details for notification: %v", err)
		return nil // Badge was awarded, notification is optional
	}

	// Send notification
	s.notifyBadgeAwarded(userID, badge)

	log.Printf("✨ Badge '%s' awarded to user %s", badgeCode, userID)
	return nil
}

// notifyBadgeAwarded sends a notification when a badge is awarded
func (s *Service) notifyBadgeAwarded(userID uuid.UUID, badge *Badge) {
	if s.hub == nil {
		log.Printf("Badge awarded to user %s: %s (no hub configured)", userID, badge.Name)
		return
	}

	notification := fmt.Sprintf(`
		<div id="notifications" hx-swap-oob="beforeend">
			<div class="alert shadow-lg mb-2 bg-gradient-to-r from-purple-500 to-pink-500 text-white animate-fade-in" role="alert">
				<div class="flex items-center gap-3">
					<span class="text-4xl">🏆</span>
					<div>
						<h3 class="font-bold text-lg">Badge Unlocked!</h3>
						<div class="text-sm"><strong>%s</strong></div>
						<div class="text-xs opacity-90 mt-1">%s</div>
					</div>
				</div>
				<button class="btn btn-sm btn-ghost text-white" onclick="this.parentElement.remove()">✕</button>
			</div>
		</div>
	`, badge.Name, badge.Description)

	s.hub.BroadcastToUser(userID.String(), []byte(notification))

	log.Printf("🏆 Badge notification sent: User %s unlocked '%s'", userID, badge.Name)
}
