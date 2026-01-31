package badges

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockNotificationHub implements NotificationHub for testing
type MockNotificationHub struct {
	Messages map[string][]string // userID -> list of messages
}

func NewMockNotificationHub() *MockNotificationHub {
	return &MockNotificationHub{
		Messages: make(map[string][]string),
	}
}

func (m *MockNotificationHub) BroadcastToUser(userID string, message []byte) {
	if m.Messages[userID] == nil {
		m.Messages[userID] = []string{}
	}
	m.Messages[userID] = append(m.Messages[userID], string(message))
}

func (m *MockNotificationHub) GetMessagesForUser(userID string) []string {
	return m.Messages[userID]
}

func (m *MockNotificationHub) Clear() {
	m.Messages = make(map[string][]string)
}

// getTestDBPool creates a connection pool to the test database
func getTestDBPool(t *testing.T) *pgxpool.Pool {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5470/myapp?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("Skipping integration test: could not connect to database: %v", err)
		return nil
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("Skipping integration test: database ping failed: %v", err)
		return nil
	}

	return pool
}

// createTestUser creates a test user for badge testing
func createTestUser(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("badge_test_%s@test.com", uniqueSuffix)
	username := fmt.Sprintf("badge_test_%s", uniqueSuffix)

	query := `
		INSERT INTO users (email, username, hashed_password, display_name, is_active)
		VALUES ($1, $2, '$2a$10$test', $3, true)
		RETURNING id
	`

	var userID uuid.UUID
	err := pool.QueryRow(ctx, query, email, username, username).Scan(&userID)
	require.NoError(t, err)

	return userID
}

// createTestUserWithCreationTime creates a test user with a specific creation time
func createTestUserWithCreationTime(t *testing.T, pool *pgxpool.Pool, createdAt time.Time) uuid.UUID {
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("badge_test_%s@test.com", uniqueSuffix)
	username := fmt.Sprintf("badge_test_%s", uniqueSuffix)

	query := `
		INSERT INTO users (email, username, hashed_password, display_name, is_active, created_at)
		VALUES ($1, $2, '$2a$10$test', $3, true, $4)
		RETURNING id
	`

	var userID uuid.UUID
	err := pool.QueryRow(ctx, query, email, username, username, createdAt).Scan(&userID)
	require.NoError(t, err)

	return userID
}

// cleanupTestUser removes a test user and their badges
func cleanupTestUser(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID) {
	ctx := context.Background()
	_, _ = pool.Exec(ctx, "DELETE FROM user_badges WHERE user_id = $1", userID)
	_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
}

// ensureFirstMatchBadgeExists ensures the first_match badge exists in the database
func ensureFirstMatchBadgeExists(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()
	query := `
		INSERT INTO badges (code, name, description, icon_url)
		VALUES ('first_match', 'First Connection', 'Successfully connected with another user.', '🤝')
		ON CONFLICT (code) DO NOTHING
	`
	_, err := pool.Exec(ctx, query)
	require.NoError(t, err)
}

// TestAwardBadge_FirstTime tests awarding a badge for the first time
func TestAwardBadge_FirstTime(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()
	ensureFirstMatchBadgeExists(t, pool)

	userID := createTestUser(t, pool)
	defer cleanupTestUser(t, pool, userID)

	mockHub := NewMockNotificationHub()
	repo := NewRepository(pool)
	service := NewService(repo, mockHub)

	// Award badge for the first time
	err := service.AwardBadge(ctx, userID, "first_match")
	require.NoError(t, err)

	// Verify badge was awarded
	badges, err := repo.GetUserBadges(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, badges, 1, "User should have exactly 1 badge")
	assert.Equal(t, "first_match", badges[0].BadgeCode)
	assert.Equal(t, "First Connection", badges[0].BadgeName)

	// Verify notification was sent
	messages := mockHub.GetMessagesForUser(userID.String())
	assert.Len(t, messages, 1, "User should have received 1 notification")
	assert.Contains(t, messages[0], "Badge Unlocked!")
	assert.Contains(t, messages[0], "First Connection")
}

// TestAwardBadge_Duplicate tests that awarding the same badge twice doesn't duplicate it
func TestAwardBadge_Duplicate(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()
	ensureFirstMatchBadgeExists(t, pool)

	userID := createTestUser(t, pool)
	defer cleanupTestUser(t, pool, userID)

	mockHub := NewMockNotificationHub()
	repo := NewRepository(pool)
	service := NewService(repo, mockHub)

	// Award badge first time
	err := service.AwardBadge(ctx, userID, "first_match")
	require.NoError(t, err)

	// Clear notifications
	mockHub.Clear()

	// Try to award same badge again
	err = service.AwardBadge(ctx, userID, "first_match")
	require.NoError(t, err, "Should not error when awarding duplicate badge")

	// Verify still only has one badge
	badges, err := repo.GetUserBadges(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, badges, 1, "User should still have only 1 badge")

	// Verify no new notification was sent
	messages := mockHub.GetMessagesForUser(userID.String())
	assert.Len(t, messages, 0, "No notification should be sent for duplicate badge")
}

// TestAwardBadge_MultipleBadges tests awarding different badges to same user
func TestAwardBadge_MultipleBadges(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()
	ensureFirstMatchBadgeExists(t, pool)

	// Ensure early_adopter badge exists
	_, err := pool.Exec(ctx, `
		INSERT INTO badges (code, name, description, icon_url)
		VALUES ('early_adopter', 'Early Adopter', 'Joined during beta.', '🚀')
		ON CONFLICT (code) DO NOTHING
	`)
	require.NoError(t, err)

	userID := createTestUser(t, pool)
	defer cleanupTestUser(t, pool, userID)

	mockHub := NewMockNotificationHub()
	repo := NewRepository(pool)
	service := NewService(repo, mockHub)

	// Award first badge
	err = service.AwardBadge(ctx, userID, "first_match")
	require.NoError(t, err)

	// Award second badge
	err = service.AwardBadge(ctx, userID, "early_adopter")
	require.NoError(t, err)

	// Verify both badges awarded
	badges, err := repo.GetUserBadges(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, badges, 2, "User should have 2 badges")

	// Verify both notifications sent
	messages := mockHub.GetMessagesForUser(userID.String())
	assert.Len(t, messages, 2, "User should have received 2 notifications")
}

// TestHasBadge tests the HasBadge repository method
func TestHasBadge(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()
	ensureFirstMatchBadgeExists(t, pool)

	userID := createTestUser(t, pool)
	defer cleanupTestUser(t, pool, userID)

	repo := NewRepository(pool)

	// Initially should not have badge
	hasBadge, err := repo.HasBadge(ctx, userID, "first_match")
	require.NoError(t, err)
	assert.False(t, hasBadge, "User should not have badge initially")

	// Award badge
	err = repo.AwardBadge(ctx, userID, "first_match")
	require.NoError(t, err)

	// Now should have badge
	hasBadge, err = repo.HasBadge(ctx, userID, "first_match")
	require.NoError(t, err)
	assert.True(t, hasBadge, "User should have badge after awarding")

	// Should not have different badge
	hasBadge, err = repo.HasBadge(ctx, userID, "early_adopter")
	require.NoError(t, err)
	assert.False(t, hasBadge, "User should not have unaward badge")
}

// TestGetBadgeByCode tests retrieving badge details
func TestGetBadgeByCode(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()
	ensureFirstMatchBadgeExists(t, pool)

	repo := NewRepository(pool)

	// Get existing badge
	badge, err := repo.GetBadgeByCode(ctx, "first_match")
	require.NoError(t, err)
	assert.NotNil(t, badge)
	assert.Equal(t, "first_match", badge.Code)
	assert.Equal(t, "First Connection", badge.Name)
	assert.Contains(t, badge.Description, "connected")

	// Try to get non-existent badge
	_, err = repo.GetBadgeByCode(ctx, "nonexistent_badge")
	assert.Error(t, err, "Should error when badge doesn't exist")
}

// TestAwardBadge_WithoutHub tests badge awarding without notification hub
func TestAwardBadge_WithoutHub(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()
	ensureFirstMatchBadgeExists(t, pool)

	userID := createTestUser(t, pool)
	defer cleanupTestUser(t, pool, userID)

	repo := NewRepository(pool)
	service := NewService(repo, nil) // No notification hub

	// Should still award badge successfully
	err := service.AwardBadge(ctx, userID, "first_match")
	require.NoError(t, err)

	// Verify badge was awarded
	badges, err := repo.GetUserBadges(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, badges, 1)
}

// TestAwardBadge_InvalidBadgeCode tests awarding non-existent badge
func TestAwardBadge_InvalidBadgeCode(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()

	userID := createTestUser(t, pool)
	defer cleanupTestUser(t, pool, userID)

	mockHub := NewMockNotificationHub()
	repo := NewRepository(pool)
	service := NewService(repo, mockHub)

	// Try to award non-existent badge
	err := service.AwardBadge(ctx, userID, "nonexistent_badge")
	// Should handle gracefully (AwardBadge in repo does INSERT ... SELECT which won't fail if badge doesn't exist)
	// The badge just won't be awarded
	assert.NoError(t, err)

	// Verify no badge was awarded
	badges, err := repo.GetUserBadges(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, badges, 0)
}

// TestGetUserBadges_OrderedByDate tests badges are returned in correct order
func TestGetUserBadges_OrderedByDate(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()
	ensureFirstMatchBadgeExists(t, pool)

	// Ensure early_adopter badge exists
	_, err := pool.Exec(ctx, `
		INSERT INTO badges (code, name, description, icon_url)
		VALUES ('early_adopter', 'Early Adopter', 'Joined during beta.', '🚀')
		ON CONFLICT (code) DO NOTHING
	`)
	require.NoError(t, err)

	userID := createTestUser(t, pool)
	defer cleanupTestUser(t, pool, userID)

	repo := NewRepository(pool)

	// Award first badge
	err = repo.AwardBadge(ctx, userID, "early_adopter")
	require.NoError(t, err)

	// Small delay to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	// Award second badge
	err = repo.AwardBadge(ctx, userID, "first_match")
	require.NoError(t, err)

	// Get badges
	badges, err := repo.GetUserBadges(ctx, userID)
	require.NoError(t, err)
	require.Len(t, badges, 2)

	// Should be ordered by awarded_at DESC (most recent first)
	assert.Equal(t, "first_match", badges[0].BadgeCode, "Most recently awarded badge should be first")
	assert.Equal(t, "early_adopter", badges[1].BadgeCode)
}

// TestBatchAwardEarlyAdopterBadges tests batch awarding logic
func TestBatchAwardEarlyAdopterBadges(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()

	// Ensure early_adopter badge exists
	_, err := pool.Exec(ctx, `
		INSERT INTO badges (code, name, description, icon_url)
		VALUES ('early_adopter', 'Early Adopter', 'Joined during beta.', '🚀')
		ON CONFLICT (code) DO NOTHING
	`)
	require.NoError(t, err)

	// Define cutoff date
	cutoffDate := time.Now().Add(-24 * time.Hour) // 1 day ago

	// Create users:
	// 1. User created BEFORE cutoff (should get badge)
	userOld := createTestUserWithCreationTime(t, pool, cutoffDate.Add(-1*time.Hour))
	defer cleanupTestUser(t, pool, userOld)

	// 2. User created AFTER cutoff (should NOT get badge)
	userNew := createTestUserWithCreationTime(t, pool, cutoffDate.Add(1*time.Hour))
	defer cleanupTestUser(t, pool, userNew)

	repo := NewRepository(pool)
	service := NewService(repo, nil) // Hub not needed for batch

	// Run batch award
	count, err := service.BatchAwardEarlyAdopterBadges(ctx, cutoffDate)
	require.NoError(t, err)

	// We expect at least 1 user (our test userOld).
	if count < 1 {
		t.Errorf("Expected at least 1 user to be awarded, got %d", count)
	}

	// Verify userOld HAS the badge
	hasBadge, err := repo.HasBadge(ctx, userOld, BadgeEarlyAdopter)
	require.NoError(t, err)
	assert.True(t, hasBadge, "Old user should have early_adopter badge")

	// Verify userNew DOES NOT HAVE the badge
	hasBadge, err = repo.HasBadge(ctx, userNew, BadgeEarlyAdopter)
	require.NoError(t, err)
	assert.False(t, hasBadge, "New user should NOT have early_adopter badge")

	// Run again (idempotency check)
	count, err = service.BatchAwardEarlyAdopterBadges(ctx, cutoffDate)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "Second run should award 0 badges")
}
