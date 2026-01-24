package report

import (
	"time"

	"github.com/google/uuid"
)

type ContentType string

const (
	ContentTypeMessage       ContentType = "message"
	ContentTypeProfile       ContentType = "profile"
	ContentTypeSkill         ContentType = "skill"
	ContentTypeSessionReview ContentType = "session_review"
)

type ReportStatus string

const (
	ReportStatusPending     ReportStatus = "pending"
	ReportStatusUnderReview ReportStatus = "under_review"
	ReportStatusResolved    ReportStatus = "resolved"
	ReportStatusDismissed   ReportStatus = "dismissed"
)

type ReportType string

const (
	ReportTypeHarassment    ReportType = "harassment"
	ReportTypeSpam          ReportType = "spam"
	ReportTypeInappropriate ReportType = "inappropriate_content"
	ReportTypeFraud         ReportType = "fraud"
	ReportTypeCopyright     ReportType = "copyright"
	ReportTypeOther         ReportType = "other"
)

type Report struct {
	ID              uuid.UUID    `json:"id" db:"id"`
	ReporterID      uuid.UUID    `json:"reporter_id" db:"reporter_id"`
	ReportedUserID  uuid.UUID    `json:"reported_user_id" db:"reported_user_id"`
	ContentID       string       `json:"content_id" db:"content_id"`
	ContentType     ContentType  `json:"content_type" db:"content_type"`
	ReportType      ReportType   `json:"report_type" db:"report_type"`
	Description     string       `json:"description" db:"description"`
	Status          ReportStatus `json:"status" db:"status"`
	AssignedAdminID *uuid.UUID   `json:"assigned_admin_id" db:"assigned_admin_id"`
	ResolutionNotes *string      `json:"resolution_notes" db:"resolution_notes"`
	CreatedAt       time.Time    `json:"created_at" db:"created_at"`
	ResolvedAt      *time.Time   `json:"resolved_at" db:"resolved_at"`

	// Joined Fields
	ReporterName     string `json:"reporter_name" db:"reporter_name"`
	ReportedUserName string `json:"reported_user_name" db:"reported_user_name"`
}
