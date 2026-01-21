-- +goose Up
CREATE TYPE dispute_status AS ENUM ('pending', 'under_review', 'resolved', 'escalated');

CREATE TABLE disputes (
                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                        session_id UUID REFERENCES user_sessions(id), -- Assuming you have a 'sessions' table

                        disputing_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                        disputed_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

                        reason TEXT NOT NULL,
                        description TEXT,

                        status dispute_status NOT NULL DEFAULT 'pending',
                        assigned_admin_id UUID REFERENCES users(id),

                        resolution TEXT,
                        winner_id UUID REFERENCES users(id),
  -- Using numeric for money is a best practice to avoid floating point inaccuracies.
                        refund_amount NUMERIC(10, 2),

                        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                        resolved_at TIMESTAMPTZ
);

CREATE INDEX idx_disputes_status ON disputes (status);
