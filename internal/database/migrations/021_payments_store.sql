-- +goose Up
-- Add a new ENUM type for the subscription source
CREATE TYPE subscription_provider AS ENUM ('stripe', 'apple_app_store', 'google_play_store');

-- Update the existing subscriptions table
ALTER TABLE subscriptions
  ADD COLUMN provider subscription_provider NOT NULL DEFAULT 'stripe', -- Add the source
      ADD COLUMN provider_subscription_id TEXT; -- To store the originalTransactionId from Apple/Google

-- Make the stripe_subscription_id nullable, as it won't exist for App Store purchases
ALTER TABLE subscriptions
  ALTER COLUMN stripe_subscription_id DROP NOT NULL;

-- Add a constraint to ensure at least one provider ID exists
ALTER TABLE subscriptions
  ADD CONSTRAINT check_provider_id
    CHECK (
      (provider = 'stripe' AND stripe_subscription_id IS NOT NULL) OR
      (provider IN ('apple_app_store', 'google_play_store') AND provider_subscription_id IS NOT NULL)
      );
