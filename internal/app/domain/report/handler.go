package report

import (
	"net/http"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/auth"
	"github.com/google/uuid"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) SubmitReport(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		_ = auth.SetFlash(w, r, "Invalid request", auth.FlashError)
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}

	reporterIDStr, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok || reporterIDStr == "" {
		_ = auth.SetFlash(w, r, "You must be logged in to report", auth.FlashError)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	reporterID, _ := uuid.Parse(reporterIDStr)

	// Form fields
	reportedUserIDStr := r.FormValue("reported_user_id")
	contentID := r.FormValue("content_id")
	contentType := ContentType(r.FormValue("content_type"))
	reportType := ReportType(r.FormValue("report_type"))
	description := r.FormValue("description")

	reportedUserID, err := uuid.Parse(reportedUserIDStr)
	if err != nil {
		_ = auth.SetFlash(w, r, "Invalid user ID", auth.FlashError)
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}

	report := &Report{
		ReporterID:     reporterID,
		ReportedUserID: reportedUserID,
		ContentID:      contentID,
		ContentType:    contentType,
		ReportType:     reportType,
		Description:    description,
		Status:         ReportStatusPending,
	}

	if err := h.repo.Create(r.Context(), report); err != nil {
		_ = auth.SetFlash(w, r, "Failed to submit report", auth.FlashError)
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}

	_ = auth.SetFlash(w, r, "Report submitted successfully", auth.FlashSuccess)
	// Redirect back to where they came from (e.g., profile page)
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}
