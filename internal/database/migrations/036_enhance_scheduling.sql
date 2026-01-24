-- +goose Up
-- Add calendar-specific fields to existing sessions table
ALTER TABLE sessions
ADD COLUMN IF NOT EXISTS recurrence_rule TEXT,  -- For recurring sessions (iCal format)
ADD COLUMN IF NOT EXISTS reminder_minutes INTEGER DEFAULT 60,  -- Reminder before start
ADD COLUMN IF NOT EXISTS location_type VARCHAR(50) DEFAULT 'virtual',  -- virtual, in-person, hybrid
ADD COLUMN IF NOT EXISTS location_details TEXT,  -- Zoom link, address, etc.
ADD COLUMN IF NOT EXISTS attendees JSONB DEFAULT '[]';  -- Additional participants (future feature)

-- Add location_type enum constraint
ALTER TABLE sessions
ADD CONSTRAINT chk_location_type CHECK (location_type IN ('virtual', 'in_person', 'hybrid'));

-- Additional indexes for calendar queries
CREATE INDEX IF NOT EXISTS idx_sessions_user_date_initiator ON sessions (initiator_id, scheduled_start DESC);

CREATE INDEX IF NOT EXISTS idx_sessions_user_date_partner ON sessions (partner_id, scheduled_start DESC);

-- Index for date range queries (calendar month view)
CREATE INDEX IF NOT EXISTS idx_sessions_date_range ON sessions (scheduled_start, scheduled_end);

-- Index for conflict detection (using GiST for range overlap queries)
CREATE INDEX IF NOT EXISTS idx_sessions_time_overlap ON sessions
  USING gist (tstzrange(scheduled_start, scheduled_end, '[)'));

-- +goose Down
DROP INDEX IF EXISTS idx_sessions_time_overlap;
DROP INDEX IF EXISTS idx_sessions_date_range;
DROP INDEX IF EXISTS idx_sessions_user_date_partner;
DROP INDEX IF EXISTS idx_sessions_user_date_initiator;

ALTER TABLE sessions DROP CONSTRAINT IF EXISTS chk_location_type;

ALTER TABLE sessions
DROP COLUMN IF EXISTS attendees,
DROP COLUMN IF EXISTS location_details,
DROP COLUMN IF EXISTS location_type,
DROP COLUMN IF EXISTS reminder_minutes,
DROP COLUMN IF EXISTS recurrence_rule;