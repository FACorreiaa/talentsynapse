-- +goose Up
-- Enum type for user roles to ensure data consistency.
CREATE TYPE user_role AS ENUM ('member', 'expert', 'moderator', 'admin');

CREATE TABLE users (
  -- Core Identity
                     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                     email VARCHAR(255) UNIQUE NOT NULL,
                     username VARCHAR(50) UNIQUE NOT NULL,
                     hashed_password TEXT NOT NULL,
                     display_name VARCHAR(100) NOT NULL,
                     avatar_url TEXT,

  -- Status & Timestamps
                     role user_role NOT NULL DEFAULT 'member',
                     is_active BOOLEAN NOT NULL DEFAULT true,
                     email_verified_at TIMESTAMPTZ, -- NULL if not verified
                     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                     updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                     last_login_at TIMESTAMPTZ
);

-- Indexes for fast lookups
CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_username ON users (username);


CREATE TABLE user_oauth_identities (
    provider_name VARCHAR(50) NOT NULL, -- e.g., 'google', 'github'
    provider_user_id VARCHAR(255) NOT NULL, -- The unique ID from the provider (e.g., Google's 'sub' claim)
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Store the access/refresh tokens from the provider if you need to make API calls on behalf of the user.
    -- Encrypt these tokens at rest for security.
    provider_access_token TEXT,
    provider_refresh_token TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- A user can only link one account per provider.
    PRIMARY KEY (provider_name, provider_user_id)
);

CREATE INDEX idx_user_oauth_user_id ON user_oauth_identities (user_id);

-- Trigger to auto-update updated_at timestamp
CREATE TRIGGER update_user_oauth_identities_updated_at
  BEFORE UPDATE ON user_oauth_identities
  FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- Trigger to auto-update updated_at timestamp for users table
CREATE TRIGGER update_users_updated_at
  BEFORE UPDATE ON users
  FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- Vector embedding support requires pgvector extension
-- Uncomment when using postgres image with pgvector support
-- ALTER TABLE users ADD COLUMN embedding VECTOR(768);
-- CREATE INDEX idx_users_embedding_cosine ON users USING IVFFLAT (embedding vector_cosine_ops);

-- +goose Down
DROP TABLE IF EXISTS user_oauth_identities CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TYPE IF EXISTS user_role CASCADE;
