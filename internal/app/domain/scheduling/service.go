package scheduling

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetMonthView returns all sessions for a given month
func (s *Service) GetMonthView(ctx context.Context, userID uuid.UUID, year int, month time.Month) ([]SessionView, error) {
	startDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	filter := CalendarFilter{
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    []string{"pending", "confirmed", "in_progress", "scheduled"},
	}

	return s.repo.GetCalendarSessions(ctx, filter)
}

// GetWeekView returns sessions for a given week
func (s *Service) GetWeekView(ctx context.Context, userID uuid.UUID, startOfWeek time.Time) ([]SessionView, error) {
	endOfWeek := startOfWeek.AddDate(0, 0, 7)

	filter := CalendarFilter{
		UserID:    userID,
		StartDate: startOfWeek,
		EndDate:   endOfWeek,
		Status:    []string{"pending", "confirmed", "in_progress", "scheduled"},
	}

	return s.repo.GetCalendarSessions(ctx, filter)
}

// GetSessionDetails retrieves a single session
func (s *Service) GetSessionDetails(ctx context.Context, sessionID, userID uuid.UUID) (*SessionView, error) {
	return s.repo.GetSessionByID(ctx, sessionID, userID)
}

// ProposeSession validates availability and creates session with reminders
func (s *Service) ProposeSession(ctx context.Context, initiatorID, partnerID uuid.UUID, start, end time.Time, details SessionDetails) (uuid.UUID, error) {
	// Validate time is in future
	if start.Before(time.Now()) {
		return uuid.Nil, errors.New("cannot schedule sessions in the past")
	}

	// Validate end is after start
	if end.Before(start) || end.Equal(start) {
		return uuid.Nil, errors.New("session end time must be after start time")
	}

	// Validate duration (max 4 hours)
	if end.Sub(start) > 4*time.Hour {
		return uuid.Nil, errors.New("session cannot be longer than 4 hours")
	}

	// Check initiator's availability
	initiatorAvailable, err := s.repo.CheckAvailability(ctx, initiatorID, start, end)
	if err != nil {
		return uuid.Nil, err
	}
	if !initiatorAvailable {
		return uuid.Nil, errors.New("you have a conflicting session at this time")
	}

	// Check partner's availability
	partnerAvailable, err := s.repo.CheckAvailability(ctx, partnerID, start, end)
	if err != nil {
		return uuid.Nil, err
	}
	if !partnerAvailable {
		return uuid.Nil, errors.New("partner has a conflicting session at this time")
	}

	// Set default reminder if not specified
	if details.ReminderMinutes == 0 {
		details.ReminderMinutes = 60 // 1 hour before
	}

	// Set default location type
	if details.LocationType == "" {
		details.LocationType = "virtual"
	}

	// Create session
	sessionID, err := s.repo.CreateSession(ctx, initiatorID, partnerID, start, end, details)
	if err != nil {
		return uuid.Nil, err
	}

	// Create reminders (log error but don't fail session creation)
	_ = s.repo.CreateReminders(ctx, sessionID, []uuid.UUID{initiatorID, partnerID}, details.ReminderMinutes)

	return sessionID, nil
}

// ConfirmSession confirms a pending session
func (s *Service) ConfirmSession(ctx context.Context, sessionID, userID uuid.UUID) error {
	// Get session to verify user is partner
	session, err := s.repo.GetSessionByID(ctx, sessionID, userID)
	if err != nil {
		return err
	}

	// Only partner can confirm
	if session.IsInitiator {
		return errors.New("only the session partner can confirm")
	}

	// Check if session is in pending status
	if session.Status != "pending" {
		return errors.New("session is not in pending status")
	}

	return s.repo.UpdateSessionStatus(ctx, sessionID, "confirmed")
}

// CancelSession cancels a session
func (s *Service) CancelSession(ctx context.Context, sessionID, userID uuid.UUID, reason string) error {
	// Verify user is part of the session
	session, err := s.repo.GetSessionByID(ctx, sessionID, userID)
	if err != nil {
		return err
	}

	// Can't cancel completed sessions
	if session.Status == "completed" || session.Status == "cancelled" {
		return errors.New("cannot cancel a completed or already cancelled session")
	}

	// Set default reason
	if reason == "" {
		reason = "Cancelled by user"
	}

	return s.repo.CancelSession(ctx, sessionID, reason)
}

// GetUpcomingSessions retrieves next upcoming sessions
func (s *Service) GetUpcomingSessions(ctx context.Context, userID uuid.UUID, limit int) ([]SessionView, error) {
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	return s.repo.GetUpcomingSessions(ctx, userID, limit)
}

// BuildCalendarMonth constructs a calendar month view
func (s *Service) BuildCalendarMonth(ctx context.Context, userID uuid.UUID, year int, month time.Month) (*CalendarMonth, error) {
	sessions, err := s.GetMonthView(ctx, userID, year, month)
	if err != nil {
		return nil, err
	}

	// Build calendar structure
	cal := &CalendarMonth{
		Year:     year,
		Month:    month,
		WeekDays: []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"},
		Days:     []CalendarDay{},
	}

	// Group sessions by day
	sessionsByDay := make(map[int][]SessionView)
	for _, session := range sessions {
		day := session.ScheduledStart.Day()
		sessionsByDay[day] = append(sessionsByDay[day], session)
	}

	// Get number of days in month
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)
	daysInMonth := lastDay.Day()

	// Build days
	today := time.Now()
	for day := 1; day <= daysInMonth; day++ {
		dayDate := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		isToday := dayDate.Year() == today.Year() &&
			dayDate.Month() == today.Month() &&
			dayDate.Day() == today.Day()

		cal.Days = append(cal.Days, CalendarDay{
			Day:      day,
			Sessions: sessionsByDay[day],
			IsToday:  isToday,
		})
	}

	return cal, nil
}
