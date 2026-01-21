-- +goose Up
-- A simple key-value store for platform-wide settings.
CREATE TABLE platform_settings (
  key VARCHAR(100) PRIMARY KEY,
  value TEXT NOT NULL,
  description TEXT,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );

-- Example rows:
-- ('maintenance_mode', 'false', 'Enable to show maintenance page to users')
-- ('new_signups_enabled', 'true', 'Disable to prevent new user registrations')


CREATE TABLE feature_flags (
                             name VARCHAR(100) PRIMARY KEY,
                             description TEXT,
                             is_enabled BOOLEAN NOT NULL DEFAULT false,
                             updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- A linking table for user-specific feature flags.
CREATE TABLE feature_flag_users (
                                  feature_name VARCHAR(100) NOT NULL REFERENCES feature_flags(name) ON DELETE CASCADE,
                                  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                  PRIMARY KEY (feature_name, user_id)
);

CREATE TABLE audit_logs (
                          id BIGSERIAL PRIMARY KEY, -- Use a big integer for high-volume logging
                          admin_id UUID NOT NULL REFERENCES users(id),
                          action VARCHAR(255) NOT NULL, -- e.g., 'user.suspend', 'content.remove'

                          target_type VARCHAR(50),
                          target_id TEXT,

  -- JSONB is perfect for storing unstructured details about the action.
                          details JSONB,

                          client_ip VARCHAR(50),
                          timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_admin_id ON audit_logs (admin_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs (timestamp);
CREATE INDEX idx_audit_logs_target ON audit_logs (target_type, target_id);

CREATE TYPE announcement_priority AS ENUM ('low', 'medium', 'high', 'critical');

CREATE TABLE announcements (
                             id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                             admin_id UUID NOT NULL REFERENCES users(id),
                             title TEXT NOT NULL,
                             content TEXT NOT NULL,
                             priority announcement_priority NOT NULL DEFAULT 'low',

                             created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                             expires_at TIMESTAMPTZ
);

-- A linking table for targeted announcements. If empty, it's for all users.
CREATE TABLE announcement_recipients (
                                       announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
                                       user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                       PRIMARY KEY (announcement_id, user_id)
);
