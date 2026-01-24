package scheduling

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/FACorreiaa/skillsphere/internal/app/domain/auth"
	calendarpages "github.com/FACorreiaa/skillsphere/internal/app/views/pages/calendar"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: Add proper origin check in production
	},
}

// Handler handles calendar and scheduling HTTP requests
type Handler struct {
	service *Service
	hub     *Hub
}

// NewHandler creates a new scheduling handler
func NewHandler(service *Service, hub *Hub) *Handler {
	return &Handler{service: service, hub: hub}
}

// Helper functions to convert domain types to view types
func toViewSession(s SessionView) calendarpages.SessionView {
	return calendarpages.SessionView{
		ID:              s.ID,
		PartnerID:       s.PartnerID,
		PartnerName:     s.PartnerName,
		PartnerAvatar:   s.PartnerAvatar,
		ScheduledStart:  s.ScheduledStart,
		ScheduledEnd:    s.ScheduledEnd,
		Status:          s.Status,
		InitiatorOffers: s.InitiatorOffers,
		PartnerOffers:   s.PartnerOffers,
		LocationType:    s.LocationType,
		LocationDetails: s.LocationDetails,
		MeetingURL:      s.MeetingURL,
		Notes:           s.Notes,
		IsInitiator:     s.IsInitiator,
	}
}

func toViewSessions(sessions []SessionView) []calendarpages.SessionView {
	result := make([]calendarpages.SessionView, len(sessions))
	for i, s := range sessions {
		result[i] = toViewSession(s)
	}
	return result
}

func toViewCalendarMonth(cal *CalendarMonth) *calendarpages.CalendarMonth {
	days := make([]calendarpages.CalendarDay, len(cal.Days))
	for i, d := range cal.Days {
		days[i] = calendarpages.CalendarDay{
			Day:      d.Day,
			Sessions: toViewSessions(d.Sessions),
			IsToday:  d.IsToday,
		}
	}
	return &calendarpages.CalendarMonth{
		Year:     cal.Year,
		Month:    cal.Month,
		Days:     days,
		WeekDays: cal.WeekDays,
	}
}

// ShowCalendar renders the main calendar page
func (h *Handler) ShowCalendar(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, err := uuid.Parse(sessionData.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Parse query params for month/year
	year := time.Now().Year()
	month := time.Now().Month()

	if y := r.URL.Query().Get("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil {
			year = parsed
		}
	}
	if m := r.URL.Query().Get("month"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed >= 1 && parsed <= 12 {
			month = time.Month(parsed)
		}
	}

	// Fetch calendar data
	calendarMonth, err := h.service.BuildCalendarMonth(r.Context(), userID, year, month)
	if err != nil {
		http.Error(w, "Failed to load calendar", http.StatusInternalServerError)
		return
	}

	// Get upcoming sessions for sidebar
	upcomingSessions, _ := h.service.GetUpcomingSessions(r.Context(), userID, 5)

	// Convert to view types
	viewCalendar := toViewCalendarMonth(calendarMonth)
	viewUpcoming := toViewSessions(upcomingSessions)

	// Render
	component := calendarpages.CalendarPage(sessionData, viewCalendar, viewUpcoming)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetCalendarMonth handles HTMX requests for month navigation
func (h *Handler) GetCalendarMonth(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(sessionData.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))

	if year == 0 || month < 1 || month > 12 {
		http.Error(w, "Invalid date", http.StatusBadRequest)
		return
	}

	calendarMonth, err := h.service.BuildCalendarMonth(r.Context(), userID, year, time.Month(month))
	if err != nil {
		http.Error(w, "Failed to load month", http.StatusInternalServerError)
		return
	}

	// Convert to view type and return just the calendar grid (HTMX fragment)
	viewCalendar := toViewCalendarMonth(calendarMonth)
	component := calendarpages.CalendarGrid(viewCalendar)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ProposeSessionForm shows the scheduling form
func (h *Handler) ProposeSessionForm(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	partnerID := r.URL.Query().Get("partner")
	if partnerID == "" {
		http.Error(w, "Partner ID required", http.StatusBadRequest)
		return
	}

	component := calendarpages.SessionForm(sessionData, partnerID)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// CreateSession handles session creation
func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(sessionData.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	partnerID, err := uuid.Parse(r.FormValue("partner_id"))
	if err != nil {
		http.Error(w, "Invalid partner ID", http.StatusBadRequest)
		return
	}

	startStr := r.FormValue("start_time")
	endStr := r.FormValue("end_time")

	start, err := time.Parse("2006-01-02T15:04", startStr)
	if err != nil {
		http.Error(w, "Invalid start time format", http.StatusBadRequest)
		return
	}

	end, err := time.Parse("2006-01-02T15:04", endStr)
	if err != nil {
		http.Error(w, "Invalid end time format", http.StatusBadRequest)
		return
	}

	details := SessionDetails{
		InitiatorOffers: r.Form["initiator_offers"],
		PartnerOffers:   r.Form["partner_offers"],
		Notes:           r.FormValue("notes"),
		LocationType:    r.FormValue("location_type"),
		LocationDetails: r.FormValue("location_details"),
		ReminderMinutes: 60, // Default
	}

	// Create session
	sessionID, err := h.service.ProposeSession(r.Context(), userID, partnerID, start, end, details)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if _, writeErr := fmt.Fprintf(w, `<div class="alert alert-error">%s</div>`, err.Error()); writeErr != nil {
			log.Printf("failed to write error response: %v", writeErr)
		}
		return
	}

	// Broadcast calendar refresh to partner via WebSocket
	if h.hub != nil {
		refreshMsg := `<div hx-get="/calendar/refresh" hx-trigger="load" hx-swap="none"></div>`
		h.hub.BroadcastToUser(partnerID.String(), []byte(refreshMsg))
	}

	// Return success with trigger to close modal and refresh calendar
	w.Header().Set("HX-Trigger", `{"sessionCreated": true, "closeModal": true}`)
	component := calendarpages.SessionCreatedSuccess(sessionID)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetSessionDetails shows session details modal
func (h *Handler) GetSessionDetails(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(sessionData.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session, err := h.service.GetSessionDetails(r.Context(), sessionID, userID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	viewSession := toViewSession(*session)
	component := calendarpages.SessionDetails(&viewSession, sessionData)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ConfirmSession confirms a pending session
func (h *Handler) ConfirmSession(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(sessionData.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	err = h.service.ConfirmSession(r.Context(), sessionID, userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if _, writeErr := fmt.Fprintf(w, `<div class="alert alert-error">%s</div>`, err.Error()); writeErr != nil {
			log.Printf("failed to write error response: %v", writeErr)
		}
		return
	}

	// Notify initiator via WebSocket
	session, _ := h.service.GetSessionDetails(r.Context(), sessionID, userID)
	if session != nil && h.hub != nil {
		notifyMsg := fmt.Sprintf(`
			<div class="toast toast-top toast-end">
				<div class="alert alert-success">
					<span>%s confirmed your session!</span>
				</div>
			</div>
		`, session.PartnerName)
		h.hub.BroadcastToUser(session.PartnerID.String(), []byte(notifyMsg))
	}

	w.Header().Set("HX-Trigger", "sessionConfirmed")
	if _, err := w.Write([]byte(`<div class="alert alert-success">Session confirmed!</div>`)); err != nil {
		log.Printf("failed to write success response: %v", err)
	}
}

// CancelSession cancels a session
func (h *Handler) CancelSession(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(sessionData.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	reason := r.FormValue("reason")

	err = h.service.CancelSession(r.Context(), sessionID, userID, reason)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if _, writeErr := fmt.Fprintf(w, `<div class="alert alert-error">%s</div>`, err.Error()); writeErr != nil {
			log.Printf("failed to write error response: %v", writeErr)
		}
		return
	}

	w.Header().Set("HX-Trigger", "sessionCancelled")
	if _, err := w.Write([]byte(`<div class="alert alert-info">Session cancelled</div>`)); err != nil {
		log.Printf("failed to write cancellation response: %v", err)
	}
}

// HandleWebSocket handles calendar WebSocket connections
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	sessionData := auth.GetSessionData(r)
	if sessionData.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create client and register with hub
	client := NewClient(h.hub, conn, sessionData.UserID)
	h.hub.register <- client

	// Start pumps
	go client.WritePump()
	go client.ReadPump()
}
