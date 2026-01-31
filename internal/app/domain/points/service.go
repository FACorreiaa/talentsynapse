package points

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

// Service handles points and tier business logic
type Service struct {
	repo *Repository
	hub  NotificationHub
}

// NewService creates a new points service
func NewService(repo *Repository, hub NotificationHub) *Service {
	return &Service{
		repo: repo,
		hub:  hub,
	}
}

// AwardPoints awards points for an action and handles tier upgrades
func (s *Service) AwardPoints(ctx context.Context, userID uuid.UUID, action PointsAction) (*TierUpgrade, error) {
	points, ok := PointsReward[action]
	if !ok {
		return nil, fmt.Errorf("unknown action: %s", action)
	}

	return s.AddPoints(ctx, userID, points, action)
}

// AwardPointsWithBonus awards points with a custom multiplier (e.g., for high ratings)
func (s *Service) AwardPointsWithBonus(ctx context.Context, userID uuid.UUID, action PointsAction, multiplier float64) (*TierUpgrade, error) {
	basePoints, ok := PointsReward[action]
	if !ok {
		return nil, fmt.Errorf("unknown action: %s", action)
	}

	points := int(float64(basePoints) * multiplier)
	if points < 1 {
		points = 1
	}

	return s.AddPoints(ctx, userID, points, action)
}

// AddPoints adds points and handles tier upgrade notifications
func (s *Service) AddPoints(ctx context.Context, userID uuid.UUID, points int, action PointsAction) (*TierUpgrade, error) {
	upgrade, err := s.repo.AddPoints(ctx, userID, points, action)
	if err != nil {
		return nil, err
	}

	// If there was a tier upgrade, send notification
	if upgrade != nil {
		s.notifyTierUpgrade(upgrade)
	}

	return upgrade, nil
}

// GetUserPoints gets a user's current points and tier
func (s *Service) GetUserPoints(ctx context.Context, userID uuid.UUID) (*UserPoints, error) {
	return s.repo.GetUserPoints(ctx, userID)
}

// GetUserTier gets just the tier for a user
func (s *Service) GetUserTier(ctx context.Context, userID uuid.UUID) (Tier, error) {
	up, err := s.repo.GetUserPoints(ctx, userID)
	if err != nil {
		return TierBronze, err
	}
	return up.Tier, nil
}

// GetProgressToNextTier returns progress information for the next tier
func (s *Service) GetProgressToNextTier(ctx context.Context, userID uuid.UUID) (currentPoints int, currentTier Tier, pointsNeeded int, nextTier Tier, err error) {
	up, err := s.repo.GetUserPoints(ctx, userID)
	if err != nil {
		return 0, TierBronze, 0, "", err
	}

	currentPoints = up.Points
	currentTier = up.Tier
	pointsNeeded = PointsToNextTier(currentPoints, currentTier)

	switch currentTier {
	case TierBronze:
		nextTier = TierSilver
	case TierSilver:
		nextTier = TierGold
	default:
		nextTier = ""
	}

	return
}

// EnsureUserStats creates user stats if they don't exist
func (s *Service) EnsureUserStats(ctx context.Context, userID uuid.UUID) error {
	return s.repo.EnsureUserStats(ctx, userID)
}

// notifyTierUpgrade sends a celebration notification for tier upgrades
func (s *Service) notifyTierUpgrade(upgrade *TierUpgrade) {
	if s.hub == nil {
		log.Printf("Tier upgrade for user %s: %s -> %s (no hub configured)",
			upgrade.UserID, upgrade.OldTier, upgrade.NewTier)
		return
	}

	tierInfo := TierInfoMap[upgrade.NewTier]

	notification := fmt.Sprintf(`
		<div id="notifications" hx-swap-oob="beforeend">
			<div class="alert shadow-lg mb-2 bg-gradient-to-r from-yellow-400 to-orange-500 text-white" role="alert">
				<div class="flex items-center gap-3">
					<span class="text-4xl">%s</span>
					<div>
						<h3 class="font-bold text-lg">🎉 Tier Upgrade!</h3>
						<div class="text-sm">Congratulations! You've reached <strong>%s</strong> tier!</div>
						<div class="text-xs opacity-90 mt-1">%s</div>
					</div>
				</div>
				<button class="btn btn-sm btn-ghost text-white" onclick="this.parentElement.remove()">✕</button>
			</div>
		</div>
	`, tierInfo.Badge, string(upgrade.NewTier), tierInfo.Description)

	s.hub.BroadcastToUser(upgrade.UserID.String(), []byte(notification))

	log.Printf("🎉 Tier upgrade notification sent: User %s upgraded from %s to %s with %d points",
		upgrade.UserID, upgrade.OldTier, upgrade.NewTier, upgrade.Points)
}

// AwardSessionPoints awards points when a session is completed
func (s *Service) AwardSessionPoints(ctx context.Context, userID uuid.UUID) (*TierUpgrade, error) {
	return s.AwardPoints(ctx, userID, ActionSessionCompleted)
}

// AwardReviewPoints awards points when a user receives a positive review (4-5 stars)
// Only positive reviews (4+ stars) award points
func (s *Service) AwardReviewPoints(ctx context.Context, userID uuid.UUID, rating int) (*TierUpgrade, error) {
	// Only award points for positive reviews (4-5 stars)
	if rating < 4 {
		return nil, nil // No points for reviews below 4 stars
	}

	// 5-star reviews get bonus points (double)
	if rating == 5 {
		return s.AwardPointsWithBonus(ctx, userID, ActionReviewReceived, 2.0)
	}

	// 4-star reviews get standard points
	return s.AwardPoints(ctx, userID, ActionReviewReceived)
}

// AwardProfileCompletePoints awards points for completing profile
func (s *Service) AwardProfileCompletePoints(ctx context.Context, userID uuid.UUID) (*TierUpgrade, error) {
	return s.AwardPoints(ctx, userID, ActionProfileComplete)
}

// AwardFirstMatchPoints awards points for first match
func (s *Service) AwardFirstMatchPoints(ctx context.Context, userID uuid.UUID) (*TierUpgrade, error) {
	return s.AwardPoints(ctx, userID, ActionFirstMatch)
}

// AwardFirstSessionPoints awards points for first session
func (s *Service) AwardFirstSessionPoints(ctx context.Context, userID uuid.UUID) (*TierUpgrade, error) {
	return s.AwardPoints(ctx, userID, ActionFirstSession)
}

// GetTierBadge returns the emoji badge for a tier
func GetTierBadge(tier Tier) string {
	info, ok := TierInfoMap[tier]
	if !ok {
		return "🥉"
	}
	return info.Badge
}

// GetTierDescription returns the description for a tier
func GetTierDescription(tier Tier) string {
	info, ok := TierInfoMap[tier]
	if !ok {
		return "Basic verified badge"
	}
	return info.Description
}
