-- +goose Up
-- Multi-Factor Authentication (MFA) table

CREATE TABLE IF NOT EXISTS user_mfa (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    secret TEXT NOT NULL, -- TOTP secret (encrypted)
    enabled BOOLEAN NOT NULL DEFAULT false,
    backup_codes TEXT[], -- Array of hashed backup codes
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP WITH TIME ZONE,

    -- Ensure one MFA config per user
    UNIQUE(user_id)
);

-- Index for faster lookups
CREATE INDEX idx_user_mfa_user_id ON user_mfa(user_id);
CREATE INDEX idx_user_mfa_enabled ON user_mfa(enabled);

-- Trigger to update updated_at
CREATE TRIGGER update_user_mfa_updated_at
    BEFORE UPDATE ON user_mfa
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TRIGGER IF EXISTS update_user_mfa_updated_at ON user_mfa;
DROP INDEX IF EXISTS idx_user_mfa_enabled;
DROP INDEX IF EXISTS idx_user_mfa_user_id;
DROP TABLE IF EXISTS user_mfa;
