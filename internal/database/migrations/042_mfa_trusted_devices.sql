-- +goose Up
-- Trusted devices for MFA (remember this device)

CREATE TABLE IF NOT EXISTS trusted_devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_token_hash TEXT NOT NULL, -- Hashed device token stored in cookie
    device_name TEXT, -- User-friendly device name (e.g., "Chrome on Windows")
    user_agent TEXT,
    ip_address VARCHAR(50),
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(device_token_hash)
);

-- Index for faster lookups
CREATE INDEX idx_trusted_devices_user_id ON trusted_devices (user_id);
CREATE INDEX idx_trusted_devices_token_hash ON trusted_devices (device_token_hash);
CREATE INDEX idx_trusted_devices_expires_at ON trusted_devices (expires_at);

-- MFA audit log for security events
CREATE TABLE IF NOT EXISTS mfa_audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL, -- 'mfa_enabled', 'mfa_disabled', 'mfa_verify_success', 'mfa_verify_failed', 'backup_code_used', 'trusted_device_added', 'trusted_device_removed'
    ip_address VARCHAR(50),
    user_agent TEXT,
    details JSONB, -- Additional event details
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for audit log queries
CREATE INDEX idx_mfa_audit_log_user_id ON mfa_audit_log (user_id);
CREATE INDEX idx_mfa_audit_log_event_type ON mfa_audit_log (event_type);
CREATE INDEX idx_mfa_audit_log_created_at ON mfa_audit_log (created_at DESC);

-- MFA rate limiting table
CREATE TABLE IF NOT EXISTS mfa_rate_limit (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    attempt_count INT NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    locked_until TIMESTAMP WITH TIME ZONE,

    UNIQUE(user_id)
);

CREATE INDEX idx_mfa_rate_limit_user_id ON mfa_rate_limit (user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_mfa_rate_limit_user_id;
DROP TABLE IF EXISTS mfa_rate_limit;

DROP INDEX IF EXISTS idx_mfa_audit_log_created_at;
DROP INDEX IF EXISTS idx_mfa_audit_log_event_type;
DROP INDEX IF EXISTS idx_mfa_audit_log_user_id;
DROP TABLE IF EXISTS mfa_audit_log;

DROP INDEX IF EXISTS idx_trusted_devices_expires_at;
DROP INDEX IF EXISTS idx_trusted_devices_token_hash;
DROP INDEX IF EXISTS idx_trusted_devices_user_id;
DROP TABLE IF EXISTS trusted_devices;
