-- +goose Up
CREATE TABLE user_verifications (
                                  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

                                  method VARCHAR(50) NOT NULL, -- e.g., 'expert_application', 'document_upload'
                                  status VARCHAR(20) NOT NULL, -- e.g., 'pending', 'approved', 'rejected'

  -- For storing a URL to a badge or certificate
                                  badge_url TEXT,

                                  submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                  reviewed_at TIMESTAMPTZ,
                                  reviewed_by_admin_id UUID REFERENCES users(id)
);

CREATE INDEX idx_user_verifications_user_id ON user_verifications (user_id);
