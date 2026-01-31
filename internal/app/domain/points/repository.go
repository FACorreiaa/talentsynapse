package points

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles points data access
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new points repository
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetUserPoints retrieves a user's current points and tier
func (r *Repository) GetUserPoints(ctx context.Context, userID uuid.UUID) (*UserPoints, error) {
	query := `
		SELECT user_id, COALESCE(points, 0), COALESCE(tier, 'Bronze'), COALESCE(last_updated_at, NOW())
		FROM user_stats
		WHERE user_id = $1
	`

	up := &UserPoints{
		UserID: userID,
	}

	var tierStr string
	err := r.pool.QueryRow(ctx, query, userID).Scan(&up.UserID, &up.Points, &tierStr, &up.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Return default values for new users
			return &UserPoints{
				UserID:    userID,
				Points:    0,
				Tier:      TierBronze,
				UpdatedAt: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get user points: %w", err)
	}

	up.Tier = Tier(tierStr)
	return up, nil
}

// AddPoints adds points to a user and returns the new total and any tier upgrade
func (r *Repository) AddPoints(ctx context.Context, userID uuid.UUID, points int, action PointsAction) (*TierUpgrade, error) {
	// Get current points and tier
	current, err := r.GetUserPoints(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current points: %w", err)
	}

	oldTier := current.Tier
	newPoints := current.Points + points
	newTier := CalculateTierFromPoints(newPoints)

	// Update user_stats with new points and tier
	query := `
		INSERT INTO user_stats (user_id, points, tier, last_updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			points = $2,
			tier = $3,
			last_updated_at = NOW()
	`

	_, err = r.pool.Exec(ctx, query, userID, newPoints, string(newTier))
	if err != nil {
		return nil, fmt.Errorf("failed to update points: %w", err)
	}

	// Log the points transaction
	logQuery := `
		INSERT INTO points_history (user_id, points_change, action, new_total, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`
	_, _ = r.pool.Exec(ctx, logQuery, userID, points, string(action), newPoints)

	// Check for tier upgrade
	if newTier != oldTier {
		return &TierUpgrade{
			UserID:     userID,
			OldTier:    oldTier,
			NewTier:    newTier,
			Points:     newPoints,
			UpgradedAt: time.Now(),
		}, nil
	}

	return nil, nil
}

// UpdateTier directly updates a user's tier (for manual adjustments)
func (r *Repository) UpdateTier(ctx context.Context, userID uuid.UUID, tier Tier) error {
	query := `
		INSERT INTO user_stats (user_id, tier, last_updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			tier = $2,
			last_updated_at = NOW()
	`

	_, err := r.pool.Exec(ctx, query, userID, string(tier))
	if err != nil {
		return fmt.Errorf("failed to update tier: %w", err)
	}
	return nil
}

// GetLeaderboard retrieves top users by points
func (r *Repository) GetLeaderboard(ctx context.Context, limit int) ([]UserPoints, error) {
	query := `
		SELECT us.user_id, us.points, us.tier, us.last_updated_at
		FROM user_stats us
		JOIN users u ON us.user_id = u.id
		WHERE u.is_active = true
		ORDER BY us.points DESC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}
	defer rows.Close()

	var results []UserPoints
	for rows.Next() {
		var up UserPoints
		var tierStr string
		if err := rows.Scan(&up.UserID, &up.Points, &tierStr, &up.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard row: %w", err)
		}
		up.Tier = Tier(tierStr)
		results = append(results, up)
	}

	return results, nil
}

// GetUsersByTier retrieves users with a specific tier
func (r *Repository) GetUsersByTier(ctx context.Context, tier Tier) ([]uuid.UUID, error) {
	query := `
		SELECT user_id FROM user_stats WHERE tier = $1
	`

	rows, err := r.pool.Query(ctx, query, string(tier))
	if err != nil {
		return nil, fmt.Errorf("failed to get users by tier: %w", err)
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan user id: %w", err)
		}
		userIDs = append(userIDs, id)
	}

	return userIDs, nil
}

// RecalculateAllTiers recalculates tiers for all users based on their points
func (r *Repository) RecalculateAllTiers(ctx context.Context) (int, error) {
	query := `
		UPDATE user_stats
		SET tier = CASE 
			WHEN points >= 500 THEN 'Gold'
			WHEN points >= 100 THEN 'Silver'
			ELSE 'Bronze'
		END,
		last_updated_at = NOW()
		WHERE tier != CASE 
			WHEN points >= 500 THEN 'Gold'
			WHEN points >= 100 THEN 'Silver'
			ELSE 'Bronze'
		END
	`

	result, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to recalculate tiers: %w", err)
	}

	return int(result.RowsAffected()), nil
}

// EnsureUserStats creates a user_stats record if it doesn't exist
func (r *Repository) EnsureUserStats(ctx context.Context, userID uuid.UUID) error {
	query := `
		INSERT INTO user_stats (user_id, points, tier, total_sessions, completed_sessions, average_rating, total_reviews, last_updated_at)
		VALUES ($1, 0, 'Bronze', 0, 0, 0.00, 0, NOW())
		ON CONFLICT (user_id) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to ensure user stats: %w", err)
	}
	return nil
}
