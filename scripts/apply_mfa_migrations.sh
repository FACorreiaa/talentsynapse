#!/bin/bash
# Script to manually apply pending migrations
# Run this from the skillsphere-pwa directory

DB_URL="postgres://postgres:postgres@localhost:5470/myapp?sslmode=disable"

echo "=== Creating MFA tables manually ==="

psql "$DB_URL" << 'EOF'
-- Trusted devices for MFA (remember this device)
CREATE TABLE IF NOT EXISTS trusted_devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_token_hash TEXT NOT NULL,
    device_name TEXT,
    user_agent TEXT,
    ip_address VARCHAR(50),
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(device_token_hash)
);

CREATE INDEX IF NOT EXISTS idx_trusted_devices_user_id ON trusted_devices (user_id);
CREATE INDEX IF NOT EXISTS idx_trusted_devices_token_hash ON trusted_devices (device_token_hash);
CREATE INDEX IF NOT EXISTS idx_trusted_devices_expires_at ON trusted_devices (expires_at);

-- MFA audit log for security events
CREATE TABLE IF NOT EXISTS mfa_audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    ip_address VARCHAR(50),
    user_agent TEXT,
    details JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_mfa_audit_log_user_id ON mfa_audit_log (user_id);
CREATE INDEX IF NOT EXISTS idx_mfa_audit_log_event_type ON mfa_audit_log (event_type);
CREATE INDEX IF NOT EXISTS idx_mfa_audit_log_created_at ON mfa_audit_log (created_at DESC);

-- MFA rate limiting table
CREATE TABLE IF NOT EXISTS mfa_rate_limit (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    attempt_count INT NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    locked_until TIMESTAMP WITH TIME ZONE,
    UNIQUE(user_id)
);

CREATE INDEX IF NOT EXISTS idx_mfa_rate_limit_user_id ON mfa_rate_limit (user_id);

-- Mark migrations as applied in goose_db_version
INSERT INTO goose_db_version (version_id, is_applied)
VALUES (41, true), (42, true)
ON CONFLICT DO NOTHING;

SELECT 'MFA tables created successfully!' as result;
EOF

echo "=== Done ==="
