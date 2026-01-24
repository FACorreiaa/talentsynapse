-- +goose Up
ALTER TABLE user_stats ADD COLUMN points INT NOT NULL DEFAULT 0;

ALTER TABLE user_stats
ADD COLUMN tier VARCHAR(50) NOT NULL DEFAULT 'Bronze';

-- Create index for leaderboards
CREATE INDEX idx_user_stats_points ON user_stats (points DESC);