package report

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, report *Report) error {
	query := `
		INSERT INTO content_reports (
			reporter_id, reported_user_id, content_id, content_type, 
			report_type, description, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	// Default status is 'pending'
	if report.Status == "" {
		report.Status = ReportStatusPending
	}

	err := r.db.QueryRow(ctx, query,
		report.ReporterID,
		report.ReportedUserID,
		report.ContentID,
		report.ContentType,
		report.ReportType,
		report.Description,
		report.Status,
	).Scan(&report.ID, &report.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create report: %w", err)
	}

	return nil
}

func (r *Repository) GetPending(ctx context.Context) ([]*Report, error) {
	query := `
		SELECT 
			r.id, r.reporter_id, r.reported_user_id, r.content_id, r.content_type,
			r.report_type, r.description, r.status, r.created_at,
			u1.display_name as reporter_name,
			u2.display_name as reported_user_name
		FROM content_reports r
		JOIN users u1 ON r.reporter_id = u1.id
		JOIN users u2 ON r.reported_user_id = u2.id
		WHERE r.status = 'pending'
		ORDER BY r.created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query reports: %w", err)
	}
	defer rows.Close()

	var reports []*Report
	for rows.Next() {
		var r Report
		if err := rows.Scan(
			&r.ID, &r.ReporterID, &r.ReportedUserID, &r.ContentID, &r.ContentType,
			&r.ReportType, &r.Description, &r.Status, &r.CreatedAt,
			&r.ReporterName, &r.ReportedUserName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan report: %w", err)
		}
		reports = append(reports, &r)
	}

	return reports, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status ReportStatus, notes string, adminID uuid.UUID) error {
	query := `
		UPDATE content_reports
		SET status = $1, resolution_notes = $2, assigned_admin_id = $3, resolved_at = $4
		WHERE id = $5
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query, status, notes, adminID, now, id)
	if err != nil {
		return fmt.Errorf("failed to update report status: %w", err)
	}

	return nil
}

// TODO: Add GetByID if needed for detailed view
