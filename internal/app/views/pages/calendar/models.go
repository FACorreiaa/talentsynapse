package calendar

import (
	"time"

	"github.com/google/uuid"
)

// SessionView represents a session for calendar display
type SessionView struct {
	ID              uuid.UUID
	PartnerID       uuid.UUID
	PartnerName     string
	PartnerAvatar   string
	ScheduledStart  time.Time
	ScheduledEnd    time.Time
	Status          string
	InitiatorOffers []string
	PartnerOffers   []string
	LocationType    string
	LocationDetails string
	MeetingURL      string
	Notes           string
	IsInitiator     bool
}

// CalendarDay represents a day in the calendar
type CalendarDay struct {
	Day      int
	Sessions []SessionView
	IsToday  bool
}

// CalendarMonth represents a month view
type CalendarMonth struct {
	Year     int
	Month    time.Month
	Days     []CalendarDay
	WeekDays []string
}
