-- +goose Up
-- Add missing enum values to session_status
ALTER TYPE session_status ADD VALUE IF NOT EXISTS 'pending';
ALTER TYPE session_status ADD VALUE IF NOT EXISTS 'confirmed';
ALTER TYPE session_status ADD VALUE IF NOT EXISTS 'disputed';

-- +goose Down
-- Cannot remove enum values in PostgreSQL
-- They will remain but won't be used
