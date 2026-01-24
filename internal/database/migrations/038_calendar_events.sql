-- +goose Up
-- Generic calendar events (personal events, availability blocks, etc.)
CREATE TYPE calendar_event_type AS ENUM ('availability', 'blocked', 'personal', 'holiday');

CREATE TABLE calendar_events (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  -- Event details
  title VARCHAR(255) NOT NULL,
  description TEXT,

  -- Timing
  start_time TIMESTAMPTZ NOT NULL,
  end_time TIMESTAMPTZ NOT NULL,
  all_day BOOLEAN DEFAULT FALSE,
  timezone VARCHAR(100) DEFAULT 'UTC',

  -- Type and metadata
  event_type calendar_event_type NOT NULL,
  color VARCHAR(7),  -- Hex color for calendar display

  -- Recurrence (iCal format)
  recurrence_rule TEXT,  -- e.g., "FREQ=WEEKLY;BYDAY=MO,WE,FR"
  recurrence_end TIMESTAMPTZ,

  -- Timestamps
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT valid_time_range CHECK (end_time > start_time)
);

-- Indexes for calendar queries
CREATE INDEX idx_calendar_events_user_time ON calendar_events (user_id, start_time DESC);
CREATE INDEX idx_calendar_events_time_range ON calendar_events (start_time, end_time);
CREATE INDEX idx_calendar_events_user_type ON calendar_events (user_id, event_type);

-- Index for conflict detection with sessions
CREATE INDEX idx_calendar_events_time_overlap ON calendar_events
  USING gist (tstzrange(start_time, end_time, '[)'));

-- Trigger for updated_at
CREATE TRIGGER update_calendar_events_updated_at
  BEFORE UPDATE ON calendar_events
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS update_calendar_events_updated_at ON calendar_events;
DROP INDEX IF EXISTS idx_calendar_events_time_overlap;
DROP INDEX IF EXISTS idx_calendar_events_user_type;
DROP INDEX IF EXISTS idx_calendar_events_time_range;
DROP INDEX IF EXISTS idx_calendar_events_user_time;

DROP TABLE IF EXISTS calendar_events;
DROP TYPE IF EXISTS calendar_event_type;
