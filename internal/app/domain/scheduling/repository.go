package scheduling

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetCalendarSessions retrieves all sessions in date range for a user
func (r *Repository) GetCalendarSessions(ctx context.Context, filter CalendarFilter) ([]SessionView, error) {
	query := `
		SELECT
			s.id, s.scheduled_start, s.scheduled_end, s.status,
			s.initiator_offers, s.partner_offers,
			s.location_type, s.location_details, s.meeting_url, s.notes,
			CASE
				WHEN s.initiator_id = $1 THEN s.partner_id
				ELSE s.initiator_id
			END as partner_id,
			CASE
				WHEN s.initiator_id = $1 THEN COALESCE(u2.display_name, u2.username)
				ELSE COALESCE(u1.display_name, u1.username)
			END as partner_name,
			CASE
				WHEN s.initiator_id = $1 THEN u2.avatar_url
				ELSE u1.avatar_url
			END as partner_avatar,
			(s.initiator_id = $1) as is_initiator
		FROM sessions s
		JOIN users u1 ON s.initiator_id = u1.id
		JOIN users u2 ON s.partner_id = u2.id
		WHERE (s.initiator_id = $1 OR s.partner_id = $1)
		  AND s.scheduled_start >= $2
		  AND s.scheduled_start < $3
		  AND s.status = ANY($4)
		ORDER BY s.scheduled_start ASC
	`

	rows, err := r.pool.Query(ctx, query,
		filter.UserID,
		filter.StartDate,
		filter.EndDate,
		filter.Status,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []SessionView
	for rows.Next() {
		var s SessionView
		err := rows.Scan(
			&s.ID, &s.ScheduledStart, &s.ScheduledEnd, &s.Status,
			&s.InitiatorOffers, &s.PartnerOffers,
			&s.LocationType, &s.LocationDetails, &s.MeetingURL, &s.Notes,
			&s.PartnerID, &s.PartnerName, &s.PartnerAvatar,
			&s.IsInitiator,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	return sessions, rows.Err()
}

// GetSessionByID retrieves a single session by ID for the current user
func (r *Repository) GetSessionByID(ctx context.Context, sessionID, userID uuid.UUID) (*SessionView, error) {
	query := `
		SELECT
			s.id, s.scheduled_start, s.scheduled_end, s.status,
			s.initiator_offers, s.partner_offers,
			s.location_type, s.location_details, s.meeting_url, s.notes,
			CASE
				WHEN s.initiator_id = $2 THEN s.partner_id
				ELSE s.initiator_id
			END as partner_id,
			CASE
				WHEN s.initiator_id = $2 THEN COALESCE(u2.display_name, u2.username)
				ELSE COALESCE(u1.display_name, u1.username)
			END as partner_name,
			CASE
				WHEN s.initiator_id = $2 THEN u2.avatar_url
				ELSE u1.avatar_url
			END as partner_avatar,
			(s.initiator_id = $2) as is_initiator
		FROM sessions s
		JOIN users u1 ON s.initiator_id = u1.id
		JOIN users u2 ON s.partner_id = u2.id
		WHERE s.id = $1
		  AND (s.initiator_id = $2 OR s.partner_id = $2)
	`

	var s SessionView
	err := r.pool.QueryRow(ctx, query, sessionID, userID).Scan(
		&s.ID, &s.ScheduledStart, &s.ScheduledEnd, &s.Status,
		&s.InitiatorOffers, &s.PartnerOffers,
		&s.LocationType, &s.LocationDetails, &s.MeetingURL, &s.Notes,
		&s.PartnerID, &s.PartnerName, &s.PartnerAvatar,
		&s.IsInitiator,
	)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// CreateSession creates a new skill exchange session
func (r *Repository) CreateSession(ctx context.Context, initiatorID, partnerID uuid.UUID, start, end time.Time, details SessionDetails) (uuid.UUID, error) {
	query := `
		INSERT INTO sessions (
			initiator_id, partner_id, scheduled_start, scheduled_end,
			initiator_offers, partner_offers, notes,
			location_type, location_details, reminder_minutes,
			status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'pending')
		RETURNING id
	`

	var sessionID uuid.UUID
	err := r.pool.QueryRow(ctx, query,
		initiatorID, partnerID, start, end,
		details.InitiatorOffers, details.PartnerOffers, details.Notes,
		details.LocationType, details.LocationDetails, details.ReminderMinutes,
	).Scan(&sessionID)

	return sessionID, err
}

// UpdateSessionStatus updates the status of a session
func (r *Repository) UpdateSessionStatus(ctx context.Context, sessionID uuid.UUID, status string) error {
	query := `UPDATE sessions SET status = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, sessionID)
	return err
}

// CancelSession cancels a session with a reason
func (r *Repository) CancelSession(ctx context.Context, sessionID uuid.UUID, reason string) error {
	query := `
		UPDATE sessions
		SET status = 'cancelled', cancellation_reason = $1
		WHERE id = $2
	`
	_, err := r.pool.Exec(ctx, query, reason, sessionID)
	return err
}

// CreateReminders creates reminders for a session
func (r *Repository) CreateReminders(ctx context.Context, sessionID uuid.UUID, userIDs []uuid.UUID, reminderMinutes int) error {
	// Get session scheduled time
	var scheduledStart time.Time
	err := r.pool.QueryRow(ctx,
		"SELECT scheduled_start FROM sessions WHERE id = $1",
		sessionID,
	).Scan(&scheduledStart)
	if err != nil {
		return err
	}

	remindAt := scheduledStart.Add(-time.Duration(reminderMinutes) * time.Minute)

	// Insert reminders for each user
	query := `
		INSERT INTO reminders (session_id, user_id, remind_at, delivery_method, message)
		SELECT $1, $2, $3,
			COALESCE(
				CASE
					WHEN unp.session_reminders THEN 'all'::delivery_method
					ELSE 'in_app'::delivery_method
				END,
				'in_app'::delivery_method
			),
			$4
		FROM (SELECT $2::uuid as uid) u
		LEFT JOIN user_notification_preferences unp ON unp.user_id = u.uid
	`

	for _, userID := range userIDs {
		message := fmt.Sprintf("Your skill exchange session starts in %d minutes", reminderMinutes)
		_, err := r.pool.Exec(ctx, query, sessionID, userID, remindAt, message)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetPendingReminders fetches reminders due for sending
func (r *Repository) GetPendingReminders(ctx context.Context) ([]Reminder, error) {
	query := `
		SELECT r.id, r.session_id, r.user_id, r.remind_at,
			   r.delivery_method, r.status, r.message
		FROM reminders r
		WHERE r.status = 'pending'
		  AND r.remind_at <= NOW()
		ORDER BY r.remind_at ASC
		LIMIT 100
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []Reminder
	for rows.Next() {
		var r Reminder
		err := rows.Scan(&r.ID, &r.SessionID, &r.UserID, &r.RemindAt,
			&r.DeliveryMethod, &r.Status, &r.Message)
		if err != nil {
			return nil, err
		}
		reminders = append(reminders, r)
	}

	return reminders, rows.Err()
}

// MarkReminderSent updates reminder status after sending
func (r *Repository) MarkReminderSent(ctx context.Context, reminderID uuid.UUID, success bool, errorMsg string) error {
	status := "sent"
	if !success {
		status = "failed"
	}

	query := `
		UPDATE reminders
		SET status = $1, sent_at = NOW(), error_message = $2
		WHERE id = $3
	`

	_, err := r.pool.Exec(ctx, query, status, errorMsg, reminderID)
	return err
}

// CheckAvailability checks if a time slot is available for a user
func (r *Repository) CheckAvailability(ctx context.Context, userID uuid.UUID, start, end time.Time) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM sessions
		WHERE (initiator_id = $1 OR partner_id = $1)
		  AND status IN ('pending', 'confirmed', 'in_progress')
		  AND (
			  (scheduled_start <= $2 AND scheduled_end > $2) OR
			  (scheduled_start < $3 AND scheduled_end >= $3) OR
			  (scheduled_start >= $2 AND scheduled_end <= $3)
		  )
	`

	var count int
	err := r.pool.QueryRow(ctx, query, userID, start, end).Scan(&count)
	return count == 0, err
}

// GetUpcomingSessions returns next N upcoming sessions for a user
func (r *Repository) GetUpcomingSessions(ctx context.Context, userID uuid.UUID, limit int) ([]SessionView, error) {
	query := `
		SELECT
			s.id, s.scheduled_start, s.scheduled_end, s.status,
			s.initiator_offers, s.partner_offers,
			s.location_type, s.location_details, s.meeting_url, s.notes,
			CASE
				WHEN s.initiator_id = $1 THEN s.partner_id
				ELSE s.initiator_id
			END as partner_id,
			CASE
				WHEN s.initiator_id = $1 THEN COALESCE(u2.display_name, u2.username)
				ELSE COALESCE(u1.display_name, u1.username)
			END as partner_name,
			CASE
				WHEN s.initiator_id = $1 THEN u2.avatar_url
				ELSE u1.avatar_url
			END as partner_avatar,
			(s.initiator_id = $1) as is_initiator
		FROM sessions s
		JOIN users u1 ON s.initiator_id = u1.id
		JOIN users u2 ON s.partner_id = u2.id
		WHERE (s.initiator_id = $1 OR s.partner_id = $1)
		  AND s.scheduled_start > NOW()
		  AND s.status IN ('pending', 'confirmed')
		ORDER BY s.scheduled_start ASC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []SessionView
	for rows.Next() {
		var s SessionView
		err := rows.Scan(
			&s.ID, &s.ScheduledStart, &s.ScheduledEnd, &s.Status,
			&s.InitiatorOffers, &s.PartnerOffers,
			&s.LocationType, &s.LocationDetails, &s.MeetingURL, &s.Notes,
			&s.PartnerID, &s.PartnerName, &s.PartnerAvatar,
			&s.IsInitiator,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	return sessions, rows.Err()
}
