package chat

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Conversation represents a chat conversation between two users
type Conversation struct {
	ID            string
	PartnerID     string
	PartnerName   string
	PartnerAvatar string
	LastMessage   string
	LastMessageAt time.Time
	UnreadCount   int
	IsOnline      bool
}

// Message represents a chat message
type Message struct {
	ID             string
	ConversationID string
	SenderID       string
	SenderName     string
	SenderAvatar   string
	Content        string
	Type           string // text, image, file, system
	SentAt         time.Time
	IsOwn          bool
}

// Repository handles database operations for chat
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new chat repository
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetConversations retrieves all conversations for a user
func (r *Repository) GetConversations(ctx context.Context, userID string) ([]Conversation, error) {
	query := `
		SELECT 
			c.id,
			CASE WHEN c.user_a_id = $1 THEN c.user_b_id ELSE c.user_a_id END as partner_id,
			CASE WHEN c.user_a_id = $1 THEN ub.display_name ELSE ua.display_name END as partner_name,
			CASE WHEN c.user_a_id = $1 THEN COALESCE(ub.avatar_url, '') ELSE COALESCE(ua.avatar_url, '') END as partner_avatar,
			COALESCE(m.content, '') as last_message,
			COALESCE(c.last_message_at, c.created_at) as last_message_at,
			COALESCE(
				(SELECT COUNT(*)::int FROM messages msg 
				 WHERE msg.conversation_id = c.id 
				 AND msg.sender_id != $1 
				 AND msg.sent_at > COALESCE(cp.last_read_at, '1970-01-01')), 0
			) as unread_count
		FROM conversations c
		JOIN users ua ON c.user_a_id = ua.id
		JOIN users ub ON c.user_b_id = ub.id
		LEFT JOIN messages m ON c.last_message_id = m.id
		LEFT JOIN conversation_participants cp ON cp.conversation_id = c.id AND cp.user_id = $1
		WHERE c.user_a_id = $1 OR c.user_b_id = $1
		ORDER BY COALESCE(c.last_message_at, c.created_at) DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		var c Conversation
		err := rows.Scan(
			&c.ID,
			&c.PartnerID,
			&c.PartnerName,
			&c.PartnerAvatar,
			&c.LastMessage,
			&c.LastMessageAt,
			&c.UnreadCount,
		)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, c)
	}

	return conversations, rows.Err()
}

// GetOrCreateConversation finds or creates a 1:1 conversation between two users
func (r *Repository) GetOrCreateConversation(ctx context.Context, userA, userB string) (string, error) {
	// Use LEAST/GREATEST to ensure consistent ordering
	query := `
		INSERT INTO conversations (user_a_id, user_b_id)
		VALUES (LEAST($1::uuid, $2::uuid), GREATEST($1::uuid, $2::uuid))
		ON CONFLICT (LEAST(user_a_id, user_b_id), GREATEST(user_a_id, user_b_id)) 
		DO UPDATE SET created_at = conversations.created_at
		RETURNING id
	`

	var id string
	err := r.pool.QueryRow(ctx, query, userA, userB).Scan(&id)
	if err != nil {
		return "", err
	}

	// Ensure both users are participants
	participantQuery := `
		INSERT INTO conversation_participants (conversation_id, user_id)
		VALUES ($1, $2), ($1, $3)
		ON CONFLICT (conversation_id, user_id) DO NOTHING
	`
	_, err = r.pool.Exec(ctx, participantQuery, id, userA, userB)
	if err != nil {
		return "", err
	}

	return id, nil
}

// GetConversation retrieves a single conversation with partner details
func (r *Repository) GetConversation(ctx context.Context, conversationID, userID string) (*Conversation, error) {
	query := `
		SELECT 
			c.id,
			CASE WHEN c.user_a_id = $2 THEN c.user_b_id ELSE c.user_a_id END as partner_id,
			CASE WHEN c.user_a_id = $2 THEN ub.display_name ELSE ua.display_name END as partner_name,
			CASE WHEN c.user_a_id = $2 THEN COALESCE(ub.avatar_url, '') ELSE COALESCE(ua.avatar_url, '') END as partner_avatar
		FROM conversations c
		JOIN users ua ON c.user_a_id = ua.id
		JOIN users ub ON c.user_b_id = ub.id
		WHERE c.id = $1
		AND (c.user_a_id = $2 OR c.user_b_id = $2)
	`

	var c Conversation
	err := r.pool.QueryRow(ctx, query, conversationID, userID).Scan(
		&c.ID,
		&c.PartnerID,
		&c.PartnerName,
		&c.PartnerAvatar,
	)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// GetMessages retrieves messages for a conversation with pagination
func (r *Repository) GetMessages(ctx context.Context, conversationID, userID string, limit int, beforeID string) ([]Message, error) {
	var query string
	var args []any

	if beforeID == "" {
		query = `
			SELECT 
				m.id,
				m.conversation_id,
				m.sender_id,
				u.display_name,
				COALESCE(u.avatar_url, '') as avatar_url,
				m.content,
				m.type::text,
				m.sent_at,
				(m.sender_id = $2) as is_own
			FROM messages m
			JOIN users u ON m.sender_id = u.id
			WHERE m.conversation_id = $1 AND m.is_deleted = false
			ORDER BY m.sent_at DESC
			LIMIT $3
		`
		args = []any{conversationID, userID, limit}
	} else {
		query = `
			SELECT 
				m.id,
				m.conversation_id,
				m.sender_id,
				u.display_name,
				COALESCE(u.avatar_url, '') as avatar_url,
				m.content,
				m.type::text,
				m.sent_at,
				(m.sender_id = $2) as is_own
			FROM messages m
			JOIN users u ON m.sender_id = u.id
			WHERE m.conversation_id = $1 
			AND m.is_deleted = false
			AND m.sent_at < (SELECT sent_at FROM messages WHERE id = $4)
			ORDER BY m.sent_at DESC
			LIMIT $3
		`
		args = []any{conversationID, userID, limit, beforeID}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		err := rows.Scan(
			&m.ID,
			&m.ConversationID,
			&m.SenderID,
			&m.SenderName,
			&m.SenderAvatar,
			&m.Content,
			&m.Type,
			&m.SentAt,
			&m.IsOwn,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	// Reverse to get chronological order (oldest first)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, rows.Err()
}

// CreateMessage stores a new message and updates conversation metadata
func (r *Repository) CreateMessage(ctx context.Context, conversationID, senderID, content, msgType string) (*Message, error) {
	if msgType == "" {
		msgType = "text"
	}

	query := `
		WITH new_message AS (
			INSERT INTO messages (conversation_id, sender_id, content, type)
			VALUES ($1, $2, $3, $4::message_type)
			RETURNING id, conversation_id, sender_id, content, type::text, sent_at
		)
		UPDATE conversations 
		SET last_message_id = (SELECT id FROM new_message),
		    last_message_at = (SELECT sent_at FROM new_message)
		WHERE id = $1
		RETURNING (SELECT id FROM new_message)
	`

	var messageID string
	err := r.pool.QueryRow(ctx, query, conversationID, senderID, content, msgType).Scan(&messageID)
	if err != nil {
		return nil, err
	}

	// Fetch the full message with sender details
	msgQuery := `
		SELECT 
			m.id, m.conversation_id, m.sender_id, u.display_name, 
			COALESCE(u.avatar_url, ''), m.content, m.type::text, m.sent_at
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.id = $1
	`

	var m Message
	err = r.pool.QueryRow(ctx, msgQuery, messageID).Scan(
		&m.ID, &m.ConversationID, &m.SenderID, &m.SenderName,
		&m.SenderAvatar, &m.Content, &m.Type, &m.SentAt,
	)
	if err != nil {
		return nil, err
	}
	m.IsOwn = true

	return &m, nil
}

// MarkAsRead updates the last_read_at timestamp for a user in a conversation
func (r *Repository) MarkAsRead(ctx context.Context, conversationID, userID string) error {
	query := `
		UPDATE conversation_participants
		SET last_read_at = NOW()
		WHERE conversation_id = $1 AND user_id = $2
	`
	_, err := r.pool.Exec(ctx, query, conversationID, userID)
	return err
}

// IsParticipant checks if a user is part of a conversation
func (r *Repository) IsParticipant(ctx context.Context, conversationID, userID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM conversations 
			WHERE id = $1 AND (user_a_id = $2 OR user_b_id = $2)
		)
	`
	var exists bool
	err := r.pool.QueryRow(ctx, query, conversationID, userID).Scan(&exists)
	return exists, err
}
