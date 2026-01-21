-- +goose Up
-- Skill exchange sessions table (not to be confused with user_sessions for authentication)

CREATE TYPE session_status AS ENUM (
  'pending',        -- Session requested but not confirmed
  'confirmed',      -- Both parties confirmed
  'in_progress',    -- Session is currently happening
  'completed',      -- Successfully completed
  'cancelled',      -- Cancelled before starting
  'no_show',        -- One or both parties didn't show up
  'disputed'        -- Under dispute
);

CREATE TABLE sessions (
                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

                        -- Participants
                        initiator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                        partner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

                        -- Skills being exchanged (stored as text arrays for flexibility)
                        initiator_offers TEXT[] NOT NULL, -- Skills initiator will teach
                        partner_offers TEXT[] NOT NULL,   -- Skills partner will teach

                        -- Scheduling
                        scheduled_start TIMESTAMPTZ NOT NULL,
                        scheduled_end TIMESTAMPTZ NOT NULL,
                        actual_start TIMESTAMPTZ,
                        actual_end TIMESTAMPTZ,

                        -- Session details
                        status session_status NOT NULL DEFAULT 'pending',
                        meeting_url TEXT,
                        notes TEXT,

                        -- Premium feature flag
                        is_premium BOOLEAN NOT NULL DEFAULT false,

                        -- Cancellation/completion details
                        cancellation_reason TEXT,

                        -- Timestamps
                        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common query patterns
CREATE INDEX idx_sessions_initiator_id ON sessions (initiator_id);
CREATE INDEX idx_sessions_partner_id ON sessions (partner_id);
CREATE INDEX idx_sessions_status ON sessions (status);
CREATE INDEX idx_sessions_scheduled_start ON sessions (scheduled_start);
CREATE INDEX idx_sessions_scheduled_end ON sessions (scheduled_end);
CREATE INDEX idx_sessions_created_at ON sessions (created_at);
CREATE INDEX idx_sessions_is_premium ON sessions (is_premium);

-- Composite indexes for common queries
CREATE INDEX idx_sessions_initiator_status ON sessions (initiator_id, status);
CREATE INDEX idx_sessions_partner_status ON sessions (partner_id, status);
CREATE INDEX idx_sessions_status_scheduled ON sessions (status, scheduled_start);

-- For finding sessions involving a specific user
CREATE INDEX idx_sessions_participants ON sessions (initiator_id, partner_id);

-- Trigger to update updated_at timestamp
CREATE TRIGGER update_sessions_updated_at
  BEFORE UPDATE ON sessions
  FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
