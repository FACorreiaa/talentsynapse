package review

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Custom errors for review validation
var (
	ErrDuplicateReview       = errors.New("you have already submitted a review for this session")
	ErrSessionNotCompleted   = errors.New("you can only review after the session is completed")
	ErrSessionNotFound       = errors.New("session not found")
	ErrNotSessionParticipant = errors.New("you are not a participant in this session")
	ErrCannotReviewSelf      = errors.New("you cannot review yourself")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ValidateReviewSubmission validates that a review can be submitted for a session.
// It checks:
// 1. Session exists and is completed
// 2. User is a participant in the session
// 3. User hasn't already reviewed this session
// 4. User is not trying to review themselves
func (r *Repository) ValidateReviewSubmission(ctx context.Context, fromUserID, toUserID, sessionID string) error {
	// Check for self-review
	if fromUserID == toUserID {
		return ErrCannotReviewSelf
	}

	// Check session exists and get its status
	var status string
	var initiatorID, partnerID string
	query := `
		SELECT status, initiator_id::text, partner_id::text 
		FROM sessions 
		WHERE id = $1
	`
	err := r.db.QueryRow(ctx, query, sessionID).Scan(&status, &initiatorID, &partnerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSessionNotFound
		}
		return fmt.Errorf("failed to fetch session: %w", err)
	}

	// Check user is a participant
	if fromUserID != initiatorID && fromUserID != partnerID {
		return ErrNotSessionParticipant
	}

	// Check the target user is the other participant
	if toUserID != initiatorID && toUserID != partnerID {
		return ErrNotSessionParticipant
	}

	// Check session is completed
	if status != "completed" {
		return ErrSessionNotCompleted
	}

	// Check for duplicate review (same user reviewing same session)
	exists, err := r.ReviewExistsForSession(ctx, fromUserID, sessionID)
	if err != nil {
		return fmt.Errorf("failed to check for existing review: %w", err)
	}
	if exists {
		return ErrDuplicateReview
	}

	return nil
}

// ReviewExistsForSession checks if a user has already submitted a review for a specific session.
// This allows one review per session per direction (A→B and B→A are separate reviews).
func (r *Repository) ReviewExistsForSession(ctx context.Context, fromUserID, sessionID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM reviews 
			WHERE requester_id = $1 AND session_id = $2
		)
	`
	var exists bool
	err := r.db.QueryRow(ctx, query, fromUserID, sessionID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// ReviewExistsForUsers checks if a user has already reviewed another user (regardless of session).
// This is a more general check that can be used to see overall review status.
func (r *Repository) ReviewExistsForUsers(ctx context.Context, fromUserID, toUserID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM reviews 
			WHERE requester_id = $1 AND reviewer_id = $2
		)
	`
	var exists bool
	err := r.db.QueryRow(ctx, query, fromUserID, toUserID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
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
// CreateSessionReview inserts a new review linked to a session.
func (r *Repository) CreateSessionReview(ctx context.Context, fromUserID, toUserID, sessionID string, rating int, comment string) error {
	// 1. Get a valid skill ID (hack or from session? For now still hack or NULL if allowed, but schema says NOT NULL)
	// We'll pick first skill from DB for MVP to satisfy constraint
	var skillID uuid.UUID
	err := r.db.QueryRow(ctx, "SELECT id FROM skills LIMIT 1").Scan(&skillID)
	if err != nil {
		return fmt.Errorf("no skills found: %w", err)
	}

	// Insert review
	query := `
        INSERT INTO reviews (
            requester_id, reviewer_id, skill_id, session_id,
            type, depth, title, price, 
            requester_rating, requester_comment, status
        ) VALUES (
            $1, $2, $3, $4,
            'peer_review', 'standard', 'Session Review', 0.00, 
            $5, $6, 'completed'
        )
    `
	_, err = r.db.Exec(ctx, query,
		fromUserID, toUserID, skillID, sessionID,
		rating, comment,
	)
	if err != nil {
		return fmt.Errorf("failed to insert review: %w", err)
	}

	// Update Stats for the target user (toUserID)
	return r.UpdateUserStats(ctx, uuid.MustParse(toUserID))
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
        SELECT $1, total, avg_rating
        FROM stats
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

// GetStats retrieves user statistics
func (r *Repository) GetStats(ctx context.Context, userID uuid.UUID) (int, float64, error) {
	query := `SELECT total_reviews, average_rating FROM user_stats WHERE user_id = $1`
	var total int
	var avg float64
	err := r.db.QueryRow(ctx, query, userID).Scan(&total, &avg)
	if err != nil {
		return 0, 0, err
	}
	return total, avg, nil
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
