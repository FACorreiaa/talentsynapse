-- +goose Up
CREATE TABLE user_notification_preferences (
                                             user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,

  -- Booleans for each notification type
                                             session_requests BOOLEAN NOT NULL DEFAULT true,
                                             session_reminders BOOLEAN NOT NULL DEFAULT true,
                                             chat_messages BOOLEAN NOT NULL DEFAULT true,
                                             platform_updates BOOLEAN NOT NULL DEFAULT true,

  -- For push notifications vs. email
  -- 'email', 'push', 'all', 'none'
                                             channel_preference VARCHAR(20) NOT NULL DEFAULT 'all',

                                             updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
