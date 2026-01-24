-- +goose Up
-- Add session_id to reviews
ALTER TABLE reviews
ADD COLUMN session_id UUID REFERENCES sessions (id) ON DELETE SET NULL;

-- Add peer_review to review_type
-- Note: adding values to ENUM inside transaction is generally safe in modern Postgres
ALTER TYPE review_type ADD VALUE IF NOT EXISTS 'peer_review';

-- We also need a way to track if a session is reviewed?
-- We can just query the reviews table.

-- Add unique constraint so one review per session per user?
-- CREATE UNIQUE INDEX idx_reviews_session_requester ON reviews(session_id, requester_id);
-- But let's leave flexible for now.