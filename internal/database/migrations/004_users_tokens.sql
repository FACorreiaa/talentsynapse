-- +goose Up
CREATE TYPE token_type AS ENUM ('password_reset', 'email_verification');

CREATE TABLE user_tokens (
                           token_hash TEXT PRIMARY KEY, -- The token itself (hashed) is the primary key
                           user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                           type token_type NOT NULL,
                           expires_at TIMESTAMPTZ NOT NULL,
                           created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_tokens_user_id ON user_tokens (user_id);
CREATE INDEX idx_user_tokens_expires_at ON user_tokens (expires_at);
