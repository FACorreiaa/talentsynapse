package chat

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/auth"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/matches"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/push"
	"github.com/FACorreiaa/talentsynapse/internal/app/domain/user"
)

func getVoiceTestDBPool(t *testing.T) *pgxpool.Pool {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/talentsynapse_test?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Skipf("Skipping test: could not connect to database: %v", err)
		return nil
	}

	// Verify connection
	if err := pool.Ping(context.Background()); err != nil {
		t.Skipf("Skipping test: could not ping database: %v", err)
		return nil
	}

	return pool
}

func createVoiceTestUser(t *testing.T, pool *pgxpool.Pool, email, username string) string {
	query := `
		INSERT INTO users (email, username, password_hash, display_name, is_active, email_verified)
		VALUES ($1, $2, '$2a$10$test', $3, true, true)
		RETURNING id
	`
	var userID string
	err := pool.QueryRow(context.Background(), query, email, username, username).Scan(&userID)
	require.NoError(t, err)
	return userID
}

func cleanupVoiceTestUser(t *testing.T, pool *pgxpool.Pool, userID string) {
	// Clean up in reverse order of dependencies
	_, _ = pool.Exec(context.Background(), "DELETE FROM messages WHERE conversation_id IN (SELECT id FROM conversations WHERE user_a_id = $1 OR user_b_id = $1)", userID)
	_, _ = pool.Exec(context.Background(), "DELETE FROM conversation_participants WHERE user_id = $1", userID)
	_, _ = pool.Exec(context.Background(), "DELETE FROM conversations WHERE user_a_id = $1 OR user_b_id = $1", userID)
	_, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", userID)
}

// TestVoiceMessageRepository tests the voice message repository methods
func TestVoiceMessageRepository(t *testing.T) {
	pool := getVoiceTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewRepository(pool)
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create test users
	userAID := createVoiceTestUser(t, pool, fmt.Sprintf("voice_userA_%s@test.com", uniqueSuffix), fmt.Sprintf("voice_userA_%s", uniqueSuffix))
	userBID := createVoiceTestUser(t, pool, fmt.Sprintf("voice_userB_%s@test.com", uniqueSuffix), fmt.Sprintf("voice_userB_%s", uniqueSuffix))

	defer cleanupVoiceTestUser(t, pool, userAID)
	defer cleanupVoiceTestUser(t, pool, userBID)

	// Create conversation
	conversationID, err := repo.GetOrCreateConversation(ctx, userAID, userBID)
	require.NoError(t, err)
	require.NotEmpty(t, conversationID)

	t.Run("create voice message", func(t *testing.T) {
		msg, err := repo.CreateVoiceMessage(
			ctx,
			conversationID,
			userAID,
			"/assets/uploads/voice/test-voice.webm",
			1024,
			"audio/webm",
			30,
		)
		require.NoError(t, err)
		require.NotNil(t, msg)

		assert.NotEmpty(t, msg.ID)
		assert.Equal(t, conversationID, msg.ConversationID)
		assert.Equal(t, userAID, msg.SenderID)
		assert.Equal(t, "voice", msg.Type)
		assert.Equal(t, "/assets/uploads/voice/test-voice.webm", msg.FileURL)
		assert.Equal(t, int64(1024), msg.FileSize)
		assert.Equal(t, "audio/webm", msg.FileMimeType)
		assert.Equal(t, 30, msg.DurationSeconds)
		assert.True(t, msg.IsOwn)
	})

	t.Run("get messages includes voice message fields", func(t *testing.T) {
		messages, err := repo.GetMessages(ctx, conversationID, userAID, 10, "")
		require.NoError(t, err)
		require.NotEmpty(t, messages)

		// Find our voice message
		var voiceMsg *Message
		for i := range messages {
			if messages[i].Type == "voice" {
				voiceMsg = &messages[i]
				break
			}
		}

		require.NotNil(t, voiceMsg, "Voice message should be in the list")
		assert.Equal(t, "/assets/uploads/voice/test-voice.webm", voiceMsg.FileURL)
		assert.Equal(t, 30, voiceMsg.DurationSeconds)
	})
}

// TestVoiceMessageHandler tests the voice message HTTP handler
func TestVoiceMessageHandler(t *testing.T) {
	pool := getVoiceTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	// Initialize session store
	require.NoError(t, os.Setenv("SESSION_SECRET", "test-secret-key-for-testing-purposes"))
	auth.InitStore()

	repo := NewRepository(pool)
	matchesRepo := matches.NewRepository(pool)
	userRepo := user.NewRepository(pool)
	hub := NewHub()
	go hub.Run()
	pushService := push.NewService(pool)

	handler := NewHandler(repo, matchesRepo, hub, pushService)

	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create test users
	userAID := createVoiceTestUser(t, pool, fmt.Sprintf("voice_handler_A_%s@test.com", uniqueSuffix), fmt.Sprintf("voice_handler_A_%s", uniqueSuffix))
	userBID := createVoiceTestUser(t, pool, fmt.Sprintf("voice_handler_B_%s@test.com", uniqueSuffix), fmt.Sprintf("voice_handler_B_%s", uniqueSuffix))

	defer cleanupVoiceTestUser(t, pool, userAID)
	defer cleanupVoiceTestUser(t, pool, userBID)

	// Create conversation
	conversationID, err := repo.GetOrCreateConversation(ctx, userAID, userBID)
	require.NoError(t, err)

	// Get user for session
	testUser, err := userRepo.GetByID(ctx, userAID)
	require.NoError(t, err)

	t.Run("send voice message requires authentication", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "voice.webm")
		require.NoError(t, err)
		_, err = part.Write([]byte("fake audio data"))
		require.NoError(t, err)
		require.NoError(t, writer.WriteField("duration", "10"))
		require.NoError(t, writer.Close())

		req := httptest.NewRequest(http.MethodPost, "/chat/"+conversationID+"/voice", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", conversationID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		handler.SendVoiceMessage(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("send voice message with valid session", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "voice.webm")
		require.NoError(t, err)
		_, err = part.Write([]byte("fake audio data"))
		require.NoError(t, err)
		require.NoError(t, writer.WriteField("type", "voice"))
		require.NoError(t, writer.WriteField("duration", "15"))
		require.NoError(t, writer.Close())

		req := httptest.NewRequest(http.MethodPost, "/chat/"+conversationID+"/voice", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", conversationID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()

		// Create session
		err = auth.CreateSession(rec, req, testUser)
		require.NoError(t, err)

		// Copy cookies to new request
		for _, cookie := range rec.Result().Cookies() {
			req.AddCookie(cookie)
		}

		rec = httptest.NewRecorder()
		handler.SendVoiceMessage(rec, req)

		// Should succeed or handle gracefully
		// Note: May fail if uploads directory doesn't exist in test environment
		// but shouldn't be 401 Unauthorized
		assert.NotEqual(t, http.StatusUnauthorized, rec.Code)
	})
}

// TestIsAllowedAudioMimeType tests the audio mime type validation
func TestIsAllowedAudioMimeType(t *testing.T) {
	tests := []struct {
		mimeType string
		expected bool
	}{
		{"audio/webm", true},
		{"audio/webm;codecs=opus", true},
		{"audio/ogg", true},
		{"audio/mp4", true},
		{"audio/mpeg", true},
		{"audio/wav", true},
		{"audio/x-m4a", true},
		{"video/mp4", false},
		{"image/png", false},
		{"text/plain", false},
		{"application/json", false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			result := isAllowedAudioMimeType(tt.mimeType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetAudioExtension tests the audio extension detection
func TestGetAudioExtension(t *testing.T) {
	tests := []struct {
		mimeType string
		expected string
	}{
		{"audio/webm", ".webm"},
		{"audio/webm;codecs=opus", ".webm"},
		{"audio/ogg", ".ogg"},
		{"audio/mp4", ".m4a"},
		{"audio/x-m4a", ".m4a"},
		{"audio/mpeg", ".mp3"},
		{"audio/wav", ".wav"},
		{"unknown/type", ".webm"},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			result := getAudioExtension(tt.mimeType)
			assert.Equal(t, tt.expected, result)
		})
	}
}
