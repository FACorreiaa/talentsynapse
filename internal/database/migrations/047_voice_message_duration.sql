-- +goose Up
-- Add duration field for voice messages

-- Add duration column for voice messages (in seconds)
ALTER TABLE messages
ADD COLUMN duration_seconds INT;

-- Update constraint to include voice messages
ALTER TABLE messages
DROP CONSTRAINT IF EXISTS valid_file_message;

ALTER TABLE messages
ADD CONSTRAINT valid_file_message
CHECK (
    (type IN ('image', 'file', 'voice') AND file_url IS NOT NULL)
    OR type NOT IN ('image', 'file', 'voice')
);

-- Index for voice messages
CREATE INDEX idx_messages_voice ON messages(type) WHERE type = 'voice';

-- +goose Down
-- Remove index
DROP INDEX IF EXISTS idx_messages_voice;

-- Restore original constraint
ALTER TABLE messages DROP CONSTRAINT IF EXISTS valid_file_message;

ALTER TABLE messages
ADD CONSTRAINT valid_file_message
CHECK (
    (type IN ('image', 'file') AND file_url IS NOT NULL AND file_name IS NOT NULL)
    OR type NOT IN ('image', 'file')
);

-- Remove duration column
ALTER TABLE messages DROP COLUMN IF EXISTS duration_seconds;
