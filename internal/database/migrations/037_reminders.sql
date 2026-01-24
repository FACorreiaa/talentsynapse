-- +goose Up
-- Reminders for sessions
CREATE TYPE reminder_status AS ENUM ('pending', 'sent', 'failed', 'cancelled');
CREATE TYPE delivery_method AS ENUM ('email', 'push', 'in_app', 'all');

CREATE TABLE reminders (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  -- Timing
  remind_at TIMESTAMPTZ NOT NULL,

  -- Delivery
  delivery_method delivery_method NOT NULL DEFAULT 'in_app',
  status reminder_status NOT NULL DEFAULT 'pending',

  -- Content
  message TEXT NOT NULL,

  -- Metadata
  sent_at TIMESTAMPTZ,
  error_message TEXT,

  -- Timestamps
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for reminder processing
CREATE INDEX idx_reminders_pending ON reminders (remind_at)
  WHERE status = 'pending';

CREATE INDEX idx_reminders_user_status ON reminders (user_id, status);
CREATE INDEX idx_reminders_session ON reminders (session_id);
CREATE INDEX idx_reminders_status ON reminders (status);

-- Prevent duplicate reminders for same session/user/time
CREATE UNIQUE INDEX idx_reminders_unique ON reminders (session_id, user_id, remind_at)
  WHERE status != 'cancelled';

-- Trigger for updated_at
CREATE TRIGGER update_reminders_updated_at
  BEFORE UPDATE ON reminders
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS update_reminders_updated_at ON reminders;
DROP INDEX IF EXISTS idx_reminders_unique;
DROP INDEX IF EXISTS idx_reminders_status;
DROP INDEX IF EXISTS idx_reminders_session;
DROP INDEX IF EXISTS idx_reminders_user_status;
DROP INDEX IF EXISTS idx_reminders_pending;

DROP TABLE IF EXISTS reminders;
DROP TYPE IF EXISTS delivery_method;
DROP TYPE IF EXISTS reminder_status;
