-- +goose Up
ALTER TABLE users ADD COLUMN social_links JSONB DEFAULT '{}';

CREATE TABLE portfolio_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    link_url VARCHAR(255),
    image_url VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_portfolio_items_user_id ON portfolio_items (user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_portfolio_items_user_id;
DROP TABLE IF EXISTS portfolio_items;
ALTER TABLE users DROP COLUMN IF EXISTS social_links;