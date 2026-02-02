-- +goose NO TRANSACTION

-- +goose Up
-- Only including extensions that come with standard PostgreSQL
-- PostGIS, TimescaleDB, and vector require special PostgreSQL images

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE EXTENSION IF NOT EXISTS citext;

CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at()
    RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Enum types for cleaner data constraints
CREATE TYPE subscription_plan_type AS ENUM (
  'free',
  'premium_monthly',
  'premium_annual'
  );

CREATE TYPE subscription_status AS ENUM (
  'active',       -- Currently paid or free plan active
  'trialing',     -- In a trial period
  'past_due',     -- Payment failed
  'canceled',     -- Canceled by user, might still be active until end_date
  'expired'       -- Subscription period ended and not renewed
  );

CREATE TYPE poi_source AS ENUM (
  'loci_ai', -- Added by our system/AI
  'openstreetmap', -- Imported from OSM
  'user_submitted',-- Submitted by a user (maybe requires verification)
  'partner'        -- From a paying partner/featured listing
  );

-- +goose Down
DROP FUNCTION IF EXISTS set_updated_at() CASCADE;
DROP TYPE IF EXISTS subscription_plan_type CASCADE;
DROP TYPE IF EXISTS subscription_status CASCADE;
DROP TYPE IF EXISTS poi_source CASCADE;
DROP EXTENSION IF EXISTS "uuid-ossp" CASCADE;
DROP EXTENSION IF EXISTS citext CASCADE;
DROP EXTENSION IF EXISTS pg_trgm CASCADE;
