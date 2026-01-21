-- +goose Up
CREATE TYPE moderation_action_type AS ENUM ('warning', 'suspension', 'ban');

CREATE TABLE user_moderation_actions (
                                       id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                       user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                       admin_id UUID NOT NULL REFERENCES users(id), -- Admin who took the action
                                       action_type moderation_action_type NOT NULL,
                                       reason TEXT NOT NULL,

  -- For suspensions, this indicates when the suspension lifts. NULL for permanent actions.
                                       expires_at TIMESTAMPTZ,

  -- To track if a ban/suspension is currently active or has been reversed.
                                       is_active BOOLEAN NOT NULL DEFAULT true,
                                       reversed_at TIMESTAMPTZ,
                                       reversed_by_admin_id UUID REFERENCES users(id),
                                       reversal_reason TEXT,

                                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for quickly finding all actions for a specific user.
CREATE INDEX idx_user_moderation_user_id ON user_moderation_actions (user_id);
-- Index for finding active actions.
CREATE INDEX idx_user_moderation_active ON user_moderation_actions (is_active);



CREATE TYPE reportable_content_type AS ENUM ('message', 'profile', 'skill', 'session_review');
CREATE TYPE report_status AS ENUM ('pending', 'under_review', 'resolved', 'dismissed');
CREATE TYPE report_type AS ENUM ('harassment', 'spam', 'inappropriate_content', 'fraud', 'copyright', 'other');

CREATE TABLE content_reports (
                               id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                               reporter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                               reported_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

                               content_id TEXT NOT NULL, -- Generic ID for the content being reported
                               content_type reportable_content_type NOT NULL,

                               report_type report_type NOT NULL,
                               description TEXT,

                               status report_status NOT NULL DEFAULT 'pending',
                               assigned_admin_id UUID REFERENCES users(id),

                               resolution_notes TEXT, -- Notes from the admin who resolved it
                               moderation_action_id UUID REFERENCES user_moderation_actions(id), -- Link to any action taken

                               created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                               reviewed_at TIMESTAMPTZ,
                               resolved_at TIMESTAMPTZ
);

CREATE INDEX idx_content_reports_status ON content_reports (status);
CREATE INDEX idx_content_reports_reported_user_id ON content_reports (reported_user_id);
