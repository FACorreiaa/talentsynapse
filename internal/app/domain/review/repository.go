package review

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

// CreateReview inserts a new review.
// For MVP, we presume the 'reviewer' (the teacher) or 'requester' (learner) can leave a review.
// The schema is a bit complex (designed for formal code reviews), but we can map "Peer Review" as:
// Requester = The User leaving the review? Or the user receiving?
// Schema: requester_rating (rating given BY requester TO reviewer?)
// Usually: Requester (Learner) rates Reviewer (Teacher).
// Let's assume for this "Peer Review" feature:
// We insert a row where `requester_id` is the person leaving the rating, and `reviewer_id` is the target?
// NO, usually strict roles.
// Let's interpret:
// IF User A reviews User B.
// We create a review record.
// We need a Skill ID. We'll pick a dummy one or latest if possible, or just insert.
func (r *Repository) CreateSimpleReview(ctx context.Context, fromUserID, toUserID uuid.UUID, rating int, comment string) error {
	// We need a skill_id because of NOT NULL constraint.
	// Hack for MVP: Get ANY skill ID from available skills or user skills.
	// Better: Make skill_id optional in schema? No, constraints exist.
	// Solution: Fetch a skill ID from the system.

	// 1. Get a valid skill ID (any)
	var skillID uuid.UUID
	err := r.db.QueryRow(ctx, "SELECT id FROM skills LIMIT 1").Scan(&skillID)
	if err != nil {
		return fmt.Errorf("no skills found to associate review with: %w", err)
	}

	// Insert review. We treat 'fromUser' as requester and 'toUser' as reviewer for simplicity of "Requester rates Reviewer" flow.
	query := `
        INSERT INTO reviews (
            requester_id, reviewer_id, skill_id, 
            type, depth, title, price, 
            requester_rating, requester_comment, status
        ) VALUES (
            $1, $2, $3, 
            'code_review', 'standard', 'Peer Review', 0.00, 
            $4, $5, 'completed'
        )
    `
	// We default type/depth/price/title because the schema requires them but our simple "Peer Review" doesn't need them.
	_, err = r.db.Exec(ctx, query,
		fromUserID, toUserID, skillID,
		rating, comment,
	)
	if err != nil {
		return fmt.Errorf("failed to insert review: %w", err)
	}

	// Update Stats for the target user (toUserID)
	return r.UpdateUserStats(ctx, toUserID)
}

func (r *Repository) UpdateUserStats(ctx context.Context, userID uuid.UUID) error {
	// Calculate new stats
	query := `
        WITH stats AS (
            SELECT 
                COUNT(*) as total,
                COALESCE(AVG(requester_rating), 0) as avg_rating
            FROM reviews
            WHERE reviewer_id = $1 AND requester_rating IS NOT NULL
        )
        INSERT INTO user_stats (user_id, total_reviews, average_rating)
        SELECT $1, total, avg_rating FROM stats
        ON CONFLICT (user_id) DO UPDATE SET
            total_reviews = EXCLUDED.total_reviews,
            average_rating = EXCLUDED.average_rating,
            last_updated_at = NOW();
    `
	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to update user stats: %w", err)
	}
	return nil
}

func (r *Repository) GetReviewsReceived(ctx context.Context, userID uuid.UUID) ([]UserWithReview, error) {
	// Fetch reviews where this user is the reviewer (received rating from requester)
	query := `
        SELECT 
            r.requester_rating, r.requester_comment, r.requested_at,
            u.display_name, u.avatar_url
        FROM reviews r
        JOIN users u ON r.requester_id = u.id
        WHERE r.reviewer_id = $1 AND r.requester_rating IS NOT NULL
        ORDER BY r.requested_at DESC
        LIMIT 10
    `

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []UserWithReview
	for rows.Next() {
		var rev UserWithReview
		var rating int
		var comment *string
		var avatar *string
		if err := rows.Scan(&rating, &comment, &rev.CreatedAt, &rev.ReviewerName, &avatar); err != nil {
			continue
		}
		rev.Rating = rating
		if comment != nil {
			rev.Comment = *comment
		}
		if avatar != nil {
			rev.ReviewerAvatar = *avatar
		}
		reviews = append(reviews, rev)
	}
	return reviews, nil
}
