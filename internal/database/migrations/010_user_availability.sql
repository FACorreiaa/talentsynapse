-- +goose Up
CREATE TABLE user_availability (
                                 user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,

  -- Storing as a JSONB object is flexible and easy to manage.
  -- Example: {"monday": ["09:00-12:00", "14:00-17:00"], "tuesday": ...}
                                 weekly_schedule JSONB,

                                 timezone VARCHAR(100) NOT NULL, -- e.g., "Europe/Berlin", "America/New_York"
                                 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
