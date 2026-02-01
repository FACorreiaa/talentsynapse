package badges

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
)

const (
	BadgeEarlyAdopter = "early_adopter"
)

// GetEarlyAdopterCutoff returns the cutoff date for early adopter badge
// Can be configured via EARLY_ADOPTER_CUTOFF env var (RFC3339 format)
// Default: 2026-03-01 (approximately 1 month from initial launch)
func GetEarlyAdopterCutoff() time.Time {
	cutoffStr := os.Getenv("EARLY_ADOPTER_CUTOFF")
	if cutoffStr != "" {
		if t, err := time.Parse(time.RFC3339, cutoffStr); err == nil {
			return t
		}
	}
	// Default cutoff: March 1, 2026
	return time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
}

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

// BatchAwardEarlyAdopterBadges awards the early_adopter badge to all users created before the cutoff.
func (s *Service) BatchAwardEarlyAdopterBadges(ctx context.Context, cutoffDate time.Time) (int64, error) {
	// 1. Ensure the badge exists in the system (idempotent check)
	// We assume migration/seed created it, but good to check or ensure here if we wanted to be robust.
	// For now, we rely on the repository method which requires the badge to exist in 'badges' table.

	// 2. Perform the batch award
	count, err := s.repo.BatchAwardBadgesByCreationDate(ctx, BadgeEarlyAdopter, cutoffDate)
	if err != nil {
		return 0, fmt.Errorf("failed to batch award early adopter badge: %w", err)
	}

	if count > 0 {
		log.Printf("🚀 Awarded 'early_adopter' badge to %d users (cutoff: %s)", count, cutoffDate.Format(time.RFC3339))
	}

	return count, nil
}

// AwardEarlyAdopterIfEligible checks if a user registered before the cutoff and awards badge if eligible.
// This is meant to be called during or shortly after user registration.
func (s *Service) AwardEarlyAdopterIfEligible(ctx context.Context, userID uuid.UUID, userCreatedAt time.Time) error {
	cutoff := GetEarlyAdopterCutoff()

	// Only award if user created before cutoff
	if userCreatedAt.Before(cutoff) {
		if err := s.AwardBadge(ctx, userID, BadgeEarlyAdopter); err != nil {
			// Log but don't fail - badge awarding is not critical to registration
			log.Printf("Failed to award early_adopter badge to user %s: %v", userID, err)
			return err
		}
		log.Printf("✨ Early adopter badge awarded to user %s (created: %s, cutoff: %s)",
			userID, userCreatedAt.Format(time.RFC3339), cutoff.Format(time.RFC3339))
	}

	return nil
}
