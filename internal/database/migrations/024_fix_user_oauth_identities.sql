-- +goose Up
-- Add missing columns to user_oauth_identities table
-- This table was missing several important columns from the original migration

-- Check if columns already exist before adding them
-- +goose StatementBegin
DO $$
BEGIN
    -- Add provider_access_token if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'user_oauth_identities'
        AND column_name = 'provider_access_token'
    ) THEN
        ALTER TABLE user_oauth_identities
        ADD COLUMN provider_access_token TEXT;
    END IF;

    -- Add provider_refresh_token if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'user_oauth_identities'
        AND column_name = 'provider_refresh_token'
    ) THEN
        ALTER TABLE user_oauth_identities
        ADD COLUMN provider_refresh_token TEXT;
    END IF;

    -- Add created_at if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'user_oauth_identities'
        AND column_name = 'created_at'
    ) THEN
        ALTER TABLE user_oauth_identities
        ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
    END IF;

    -- Add updated_at if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'user_oauth_identities'
        AND column_name = 'updated_at'
    ) THEN
        ALTER TABLE user_oauth_identities
        ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
    END IF;
END $$;
-- +goose StatementEnd

-- Add primary key constraint if it doesn't exist
-- +goose StatementBegin
DO $$
BEGIN
    -- Check if the primary key already exists
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_name = 'user_oauth_identities'
        AND constraint_type = 'PRIMARY KEY'
    ) THEN
        -- Add the primary key
        ALTER TABLE user_oauth_identities
        ADD PRIMARY KEY (provider_name, provider_user_id);
    END IF;
END $$;
-- +goose StatementEnd

-- Create trigger for auto-updating updated_at if it doesn't exist
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.triggers
        WHERE trigger_name = 'update_user_oauth_identities_updated_at'
        AND event_object_table = 'user_oauth_identities'
    ) THEN
        CREATE TRIGGER update_user_oauth_identities_updated_at
          BEFORE UPDATE ON user_oauth_identities
          FOR EACH ROW
        EXECUTE FUNCTION set_updated_at();
    END IF;
END $$;
-- +goose StatementEnd

-- Create trigger for users table if it doesn't exist
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.triggers
        WHERE trigger_name = 'update_users_updated_at'
        AND event_object_table = 'users'
    ) THEN
        CREATE TRIGGER update_users_updated_at
          BEFORE UPDATE ON users
          FOR EACH ROW
        EXECUTE FUNCTION set_updated_at();
    END IF;
END $$;
-- +goose StatementEnd

-- Now add the indexes that depend on these columns
CREATE INDEX IF NOT EXISTS idx_user_oauth_updated_at ON user_oauth_identities (updated_at);
CREATE INDEX IF NOT EXISTS idx_user_oauth_provider_lookup ON user_oauth_identities (provider_name, provider_user_id);

-- Add comment to clarify purpose
COMMENT ON TABLE user_oauth_identities IS 'OAuth provider identities linked to user accounts. Stores tokens for API access on behalf of users.';
