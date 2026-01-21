-- +goose Up
CREATE TABLE user_stats (
                          user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,

                          total_sessions INT NOT NULL DEFAULT 0,
                          completed_sessions INT NOT NULL DEFAULT 0,
                          average_rating NUMERIC(3, 2) NOT NULL DEFAULT 0.00,
                          total_reviews INT NOT NULL DEFAULT 0,

  -- This would be updated by a background job or trigger.
                          last_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
