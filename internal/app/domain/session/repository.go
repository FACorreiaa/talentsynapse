package session

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// CreateRequest creates a new session request
func (r *Repository) CreateRequest(ctx context.Context, initiatorID, partnerID string, startTime time.Time) error {
	endTime := startTime.Add(time.Hour)
	query := `
		INSERT INTO sessions (
			initiator_id, partner_id, 
			initiator_offers, partner_offers,
			scheduled_start, scheduled_end,
			status
		) VALUES (
			$1, $2, 
			ARRAY[]::TEXT[], ARRAY[]::TEXT[], 
			$3, $4,
			'pending'
		)
	`
	_, err := r.pool.Exec(ctx, query, initiatorID, partnerID, startTime, endTime)
	return err
}

// GetUserSessions fetches sessions for a user (both as initiator and partner)
// Only returns sessions where both users have mutually accepted each other
func (r *Repository) GetUserSessions(ctx context.Context, userID string) ([]Session, error) {
	query := `
		SELECT 
			s.id, 
			s.initiator_id, s.partner_id, s.status, s.scheduled_start,
			u1.display_name as initiator_name, COALESCE(u1.avatar_url, '') as initiator_avatar,
			u2.display_name as partner_name, COALESCE(u2.avatar_url, '') as partner_avatar,
			EXISTS(SELECT 1 FROM reviews rv WHERE rv.session_id = s.id AND rv.requester_id = $1) as is_reviewed
		FROM sessions s
		JOIN users u1 ON s.initiator_id = u1.id
		JOIN users u2 ON s.partner_id = u2.id
		WHERE (s.initiator_id = $1 OR s.partner_id = $1)
		-- Only show sessions where users are mutually connected
		AND EXISTS (
			SELECT 1 
			FROM match_history mh1
			JOIN match_history mh2 ON mh1.user_id_a = mh2.user_id_b AND mh1.user_id_b = mh2.user_id_a
			WHERE mh1.user_id_a = s.initiator_id AND mh1.user_id_b = s.partner_id
			AND mh1.interaction_initiated = true
			AND mh2.interaction_initiated = true
		)
		ORDER BY s.scheduled_start DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		err := rows.Scan(
			&s.ID,
			&s.InitiatorID, &s.PartnerID, &s.Status, &s.ScheduledStart,
			&s.InitiatorName, &s.InitiatorAvatar,
			&s.PartnerName, &s.PartnerAvatar,
			&s.IsReviewed,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

// MarkCompleted updates status to completed
func (r *Repository) MarkCompleted(ctx context.Context, sessionID, userID string) error {
	// Only allow if user is part of session
	query := `
		UPDATE sessions 
		SET status = 'completed', actual_end = NOW()
		WHERE id = $1 AND (initiator_id = $2 OR partner_id = $2)
	`
	_, err := r.pool.Exec(ctx, query, sessionID, userID)
	return err
}

// GetSessionByID fetches a single session
func (r *Repository) GetSessionByID(ctx context.Context, sessionID string) (*Session, error) {
	query := `
		SELECT 
			s.id, 
			s.initiator_id, s.partner_id, s.status
		FROM sessions s
		WHERE s.id = $1
    `
	var s Session
	err := r.pool.QueryRow(ctx, query, sessionID).Scan(&s.ID, &s.InitiatorID, &s.PartnerID, &s.Status)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// CountCompletedSessions returns the number of completed sessions for a user
func (r *Repository) CountCompletedSessions(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT COUNT(*) FROM sessions 
		WHERE status = 'completed' AND (initiator_id = $1 OR partner_id = $1)
	`
	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}
