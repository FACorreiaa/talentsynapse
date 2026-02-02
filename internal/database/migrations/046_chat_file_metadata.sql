-- +goose Up
-- Add file metadata fields for image/file sharing in chat messages

-- Add file metadata columns (idempotent)
ALTER TABLE messages
ADD COLUMN IF NOT EXISTS file_url TEXT,
ADD COLUMN IF NOT EXISTS file_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS file_size BIGINT,
ADD COLUMN IF NOT EXISTS file_mime_type VARCHAR(100),
ADD COLUMN IF NOT EXISTS thumbnail_url TEXT;

-- Index for file messages (only if not exists)
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_messages_type') THEN
        CREATE INDEX idx_messages_type ON messages(type) WHERE type IN ('image', 'file');
    END IF;
END $$;
-- +goose StatementEnd

-- Add constraint to ensure file messages have required metadata (idempotent)
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'valid_file_message') THEN
        ALTER TABLE messages
        ADD CONSTRAINT valid_file_message
        CHECK (
            (type IN ('image', 'file') AND file_url IS NOT NULL AND file_name IS NOT NULL)
            OR type NOT IN ('image', 'file')
        );
    END IF;
END $$;
-- +goose StatementEnd

-- Add voice message type to enum
ALTER TYPE message_type ADD VALUE IF NOT EXISTS 'voice';

-- +goose Down
-- Remove constraint
ALTER TABLE messages DROP CONSTRAINT IF EXISTS valid_file_message;

-- Remove index
DROP INDEX IF EXISTS idx_messages_type;

-- Remove columns
ALTER TABLE messages
DROP COLUMN IF EXISTS thumbnail_url,
DROP COLUMN IF EXISTS file_mime_type,
DROP COLUMN IF EXISTS file_size,
DROP COLUMN IF EXISTS file_name,
DROP COLUMN IF EXISTS file_url;

-- Note: Cannot easily remove enum value in PostgreSQL
-- The 'voice' type will remain in the enum
