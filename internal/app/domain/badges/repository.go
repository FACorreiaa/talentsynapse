package badges

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetUserBadges fetches all badges awarded to a user
func (r *Repository) GetUserBadges(ctx context.Context, userID uuid.UUID) ([]*UserBadge, error) {
	query := `
		SELECT 
            ub.id, ub.user_id, ub.badge_id, ub.awarded_at,
            b.code, b.name, b.icon_url
		FROM user_badges ub
		JOIN badges b ON ub.badge_id = b.id
		WHERE ub.user_id = $1
		ORDER BY ub.awarded_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user badges: %w", err)
	}
	defer rows.Close()

	var badgeList []*UserBadge
	for rows.Next() {
		var b UserBadge
		if err := rows.Scan(
			&b.ID, &b.UserID, &b.BadgeID, &b.AwardedAt,
			&b.BadgeCode, &b.BadgeName, &b.BadgeIcon,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user badge: %w", err)
		}
		badgeList = append(badgeList, &b)
	}

	return badgeList, nil
}

// AwardBadge grants a badge to a user by its unique code.
// Safe to call multiple times; ignores duplicates via INSERT/ON CONFLICT or check.
func (r *Repository) AwardBadge(ctx context.Context, userID uuid.UUID, badgeCode string) error {
	// Subquery to get badge ID from code
	query := `
        INSERT INTO user_badges (user_id, badge_id)
        SELECT $1, id FROM badges WHERE code = $2
        ON CONFLICT (user_id, badge_id) DO NOTHING
    `
	_, err := r.db.Exec(ctx, query, userID, badgeCode)
	if err != nil {
		return fmt.Errorf("failed to award badge %s: %w", badgeCode, err)
	}
	return nil
}

// GetAllBadges fetches all available badges
func (r *Repository) GetAllBadges(ctx context.Context) ([]*Badge, error) {
	query := `SELECT id, code, name, description, icon_url, created_at FROM badges`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var badges []*Badge
	for rows.Next() {
		var b Badge
		if err := rows.Scan(&b.ID, &b.Code, &b.Name, &b.Description, &b.IconURL, &b.CreatedAt); err != nil {
			return nil, err
		}
		badges = append(badges, &b)
	}
	return badges, nil
}
