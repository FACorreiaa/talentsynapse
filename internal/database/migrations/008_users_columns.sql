-- +goose Up
-- This table should already exist from your AuthService schema.
-- We'll add a few columns to it.

ALTER TABLE users ADD COLUMN bio TEXT;
ALTER TABLE users ADD COLUMN location_text VARCHAR(255); -- For display purposes
-- Geometry support requires PostGIS extension
-- ALTER TABLE users ADD COLUMN location_geom GEOMETRY(Point, 4326);
-- CREATE INDEX idx_users_location_geom ON users USING GIST (location_geom);
ALTER TABLE users ADD COLUMN latitude DOUBLE PRECISION;
ALTER TABLE users ADD COLUMN longitude DOUBLE PRECISION;
ALTER TABLE users ADD COLUMN is_verified BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMPTZ; -- For soft deletes

-- Create a compound index for lat/lng queries
CREATE INDEX idx_users_location_coords ON users (latitude, longitude) WHERE latitude IS NOT NULL AND longitude IS NOT NULL;
