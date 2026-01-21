-- +goose Up
CREATE TABLE user_sessions (
                             id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                             user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  -- The refresh token should be hashed before storing, just like a password.
                             hashed_refresh_token TEXT UNIQUE NOT NULL,

  -- Metadata for user experience
                             user_agent TEXT,
                             client_ip VARCHAR(50),

                             expires_at TIMESTAMPTZ NOT NULL,
                             created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_sessions_user_id ON user_sessions (user_id);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions (expires_at);
