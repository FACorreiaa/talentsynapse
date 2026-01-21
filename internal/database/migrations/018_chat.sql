-- +goose Up
CREATE TABLE conversations (
                             id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

  -- For 1:1 chats, you can have two columns. For group chats, a linking table is better.
  -- Let's assume 1:1 for the MVP as per your proto.
                             user_a_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                             user_b_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  -- Denormalized data for fast conversation list rendering
                             last_message_id UUID, -- Can be a foreign key later
                             last_message_at TIMESTAMPTZ,

                             created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- To quickly find a conversation between two users without worrying about order (A->B vs B->A)
CREATE UNIQUE INDEX idx_conversations_unique_pair ON conversations (LEAST(user_a_id, user_b_id), GREATEST(user_a_id, user_b_id));

-- Index for finding all conversations a user is part of.
CREATE INDEX idx_conversations_user_a ON conversations (user_a_id);
CREATE INDEX idx_conversations_user_b ON conversations (user_b_id);

CREATE TYPE message_type AS ENUM ('text', 'image', 'file', 'system');

CREATE TABLE messages (
                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                        conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
                        sender_id UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,

                        type message_type NOT NULL DEFAULT 'text',
                        content TEXT NOT NULL, -- The text message or a URL to a file in object storage

                        sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  -- For soft deletes, allowing users to "delete for me"
                        is_deleted BOOLEAN NOT NULL DEFAULT false
);

-- This is the MOST IMPORTANT index for a chat application.
-- It allows for extremely fast loading of the most recent messages for a given conversation.
CREATE INDEX idx_messages_conversation_chronological ON messages (conversation_id, sent_at DESC);

CREATE INDEX idx_messages_sender_id ON messages (sender_id);

CREATE TABLE message_read_status (
                                   message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
                                   user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- The user who read the message
                                   read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                                   PRIMARY KEY (message_id, user_id)
);

CREATE TABLE conversation_participants (
                                         conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
                                         user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  -- The timestamp of the last message this user has read in this conversation.
                                         last_read_at TIMESTAMPTZ,

                                         is_archived BOOLEAN NOT NULL DEFAULT false,

                                         PRIMARY KEY (conversation_id, user_id)
);

CREATE TABLE message_reactions (
                                 message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
                                 user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                 emoji TEXT NOT NULL,

                                 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                                 PRIMARY KEY (message_id, user_id, emoji)
);

