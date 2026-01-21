-- +goose Up
-- Fix incorrect foreign key references to sessions table
-- Both disputes and match_history were referencing user_sessions(id) instead of sessions(id)

-- Fix disputes table foreign key
-- Note: This assumes disputes.session_id was meant to reference skill exchange sessions
-- The column currently references user_sessions which is for authentication
-- We need to drop and recreate the FK constraint

-- First, check if there are any existing disputes with session_ids
-- In production, you'd want to migrate or clean this data first
-- For now, we'll allow NULL session_ids temporarily

ALTER TABLE disputes
  ALTER COLUMN session_id DROP NOT NULL;

-- Drop the old incorrect foreign key if it exists
-- Note: The original migration didn't have an explicit FK name, so PostgreSQL auto-generated one
-- +goose StatementBegin
DO $$
BEGIN
    -- Try to drop the constraint if it exists
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'disputes_session_id_fkey'
        AND table_name = 'disputes'
    ) THEN
        ALTER TABLE disputes DROP CONSTRAINT disputes_session_id_fkey;
    END IF;
END $$;
-- +goose StatementEnd

-- Add the correct foreign key to sessions table
ALTER TABLE disputes
  ADD CONSTRAINT fk_disputes_session
  FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE SET NULL;

-- Fix match_history table foreign key
-- Same issue - it was referencing user_sessions instead of sessions

ALTER TABLE match_history
  ALTER COLUMN session_id DROP NOT NULL;

-- +goose StatementBegin
DO $$
BEGIN
    -- Try to drop the constraint if it exists
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'match_history_session_id_fkey'
        AND table_name = 'match_history'
    ) THEN
        ALTER TABLE match_history DROP CONSTRAINT match_history_session_id_fkey;
    END IF;
END $$;
-- +goose StatementEnd

-- Add the correct foreign key to sessions table
ALTER TABLE match_history
  ADD CONSTRAINT fk_match_history_session
  FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE SET NULL;

-- Add comment to clarify the distinction
COMMENT ON TABLE sessions IS 'Skill exchange sessions between users (not to be confused with user_sessions which is for authentication)';
COMMENT ON TABLE user_sessions IS 'Authentication session tokens (not to be confused with sessions which are skill exchanges)';
