package scheduling

import (
	"time"

	"github.com/google/uuid"
)

// SessionView combines session with user details for calendar display
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
	IsInitiator     bool // True if current user is initiator
}

// CalendarEvent for non-session events
type CalendarEvent struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Title          string
	Description    string
	StartTime      time.Time
	EndTime        time.Time
	AllDay         bool
	EventType      string // availability, blocked, personal, holiday
	Color          string
	RecurrenceRule string
	Timezone       string
}

// Reminder model
type Reminder struct {
	ID             uuid.UUID
	SessionID      uuid.UUID
	UserID         uuid.UUID
	RemindAt       time.Time
	DeliveryMethod string
	Status         string
	Message        string
	SentAt         *time.Time
	ErrorMessage   string
}

// CalendarFilter for querying sessions
type CalendarFilter struct {
	UserID    uuid.UUID
	StartDate time.Time
	EndDate   time.Time
	Status    []string // Filter by session status
}

// TimeSlot for availability checking
type TimeSlot struct {
	Start     time.Time
	End       time.Time
	Available bool
}

// SessionDetails for creating new sessions
type SessionDetails struct {
	InitiatorOffers []string
	PartnerOffers   []string
	Notes           string
	LocationType    string
	LocationDetails string
	ReminderMinutes int
}

// CalendarDay represents a day in the calendar with its sessions
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
