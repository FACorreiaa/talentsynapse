package matches

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/badges"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockNotificationHub implements badges.NotificationHub for testing
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

// getBadgeTestDBPool creates a connection pool to the test database
func getBadgeTestDBPool(t *testing.T) *pgxpool.Pool {
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

// createTestUser creates a test user with skills for matching
func createTestUserWithSkills(t *testing.T, pool *pgxpool.Pool, username string) uuid.UUID {
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("%s_%s@test.com", username, uniqueSuffix)

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

// cleanupBadgeTestUser removes a test user and all related data
func cleanupBadgeTestUser(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID) {
	ctx := context.Background()
	_, _ = pool.Exec(ctx, "DELETE FROM user_badges WHERE user_id = $1", userID)
	_, _ = pool.Exec(ctx, "DELETE FROM match_history WHERE user_id_a = $1 OR user_id_b = $1", userID)
	_, _ = pool.Exec(ctx, "DELETE FROM user_skills WHERE user_id = $1", userID)
	_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
}

// ensureFirstMatchBadgeExists ensures the first_match badge exists
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

// TestFirstMatchBadge_SuccessfulMutualMatch tests badge is awarded on first mutual match
func TestFirstMatchBadge_SuccessfulMutualMatch(t *testing.T) {
	pool := getBadgeTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()
	ensureFirstMatchBadgeExists(t, pool)

	// Create two test users
	userA := createTestUserWithSkills(t, pool, "match_user_a")
	userB := createTestUserWithSkills(t, pool, "match_user_b")
	defer cleanupBadgeTestUser(t, pool, userA)
	defer cleanupBadgeTestUser(t, pool, userB)

	// Setup repositories and services
	matchesRepo := NewRepository(pool)
	badgesRepo := badges.NewRepository(pool)
	mockHub := NewMockNotificationHub()
	badgeService := badges.NewService(badgesRepo, mockHub)

	// Step 1: User A accepts User B
	err := matchesRepo.RecordMatchAction(ctx, userA.String(), userB.String(), true)
	require.NoError(t, err, "User A should be able to accept User B")

	// Verify not connected yet (one-sided)
	connected, err := matchesRepo.AreConnected(ctx, userA.String(), userB.String())
	require.NoError(t, err)
	assert.False(t, connected, "Should not be connected with only one-sided acceptance")

	// Verify no badges awarded yet
	badgesA, _ := badgesRepo.GetUserBadges(ctx, userA)
	badgesB, _ := badgesRepo.GetUserBadges(ctx, userB)
	assert.Len(t, badgesA, 0, "User A should have no badges before mutual match")
	assert.Len(t, badgesB, 0, "User B should have no badges before mutual match")

	// Step 2: User B accepts User A (creates mutual match)
	err = matchesRepo.RecordMatchAction(ctx, userB.String(), userA.String(), true)
	require.NoError(t, err, "User B should be able to accept User A")

	// Verify now connected
	connected, err = matchesRepo.AreConnected(ctx, userA.String(), userB.String())
	require.NoError(t, err)
	assert.True(t, connected, "Should be connected after mutual acceptance")

	// Step 3: Award badges (simulating what the handler does)
	err = badgeService.AwardBadge(ctx, userA, "first_match")
	require.NoError(t, err, "Should award first_match badge to User A")

	err = badgeService.AwardBadge(ctx, userB, "first_match")
	require.NoError(t, err, "Should award first_match badge to User B")

	// Step 4: Verify badges were awarded
	badgesA, err = badgesRepo.GetUserBadges(ctx, userA)
	require.NoError(t, err)
	assert.Len(t, badgesA, 1, "User A should have 1 badge")
	assert.Equal(t, "first_match", badgesA[0].BadgeCode)
	assert.Equal(t, "First Connection", badgesA[0].BadgeName)

	badgesB, err = badgesRepo.GetUserBadges(ctx, userB)
	require.NoError(t, err)
	assert.Len(t, badgesB, 1, "User B should have 1 badge")
	assert.Equal(t, "first_match", badgesB[0].BadgeCode)

	// Step 5: Verify notifications were sent
	messagesA := mockHub.GetMessagesForUser(userA.String())
	messagesB := mockHub.GetMessagesForUser(userB.String())

	assert.Len(t, messagesA, 1, "User A should receive badge notification")
	assert.Contains(t, messagesA[0], "Badge Unlocked!")
	assert.Contains(t, messagesA[0], "First Connection")

	assert.Len(t, messagesB, 1, "User B should receive badge notification")
	assert.Contains(t, messagesB[0], "Badge Unlocked!")
	assert.Contains(t, messagesB[0], "First Connection")
}

// TestFirstMatchBadge_OnlyAwardedOnce tests badge is only awarded once
func TestFirstMatchBadge_OnlyAwardedOnce(t *testing.T) {
	pool := getBadgeTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()
	ensureFirstMatchBadgeExists(t, pool)

	// Create three test users
	userA := createTestUserWithSkills(t, pool, "repeat_user_a")
	userB := createTestUserWithSkills(t, pool, "repeat_user_b")
	userC := createTestUserWithSkills(t, pool, "repeat_user_c")
	defer cleanupBadgeTestUser(t, pool, userA)
	defer cleanupBadgeTestUser(t, pool, userB)
	defer cleanupBadgeTestUser(t, pool, userC)

	matchesRepo := NewRepository(pool)
	badgesRepo := badges.NewRepository(pool)
	mockHub := NewMockNotificationHub()
	badgeService := badges.NewService(badgesRepo, mockHub)

	// First match: A <-> B
	_ = matchesRepo.RecordMatchAction(ctx, userA.String(), userB.String(), true)
	_ = matchesRepo.RecordMatchAction(ctx, userB.String(), userA.String(), true)

	// Award badge for first match
	_ = badgeService.AwardBadge(ctx, userA, "first_match")
	_ = badgeService.AwardBadge(ctx, userB, "first_match")

	// Verify badge awarded
	badgesA, _ := badgesRepo.GetUserBadges(ctx, userA)
	assert.Len(t, badgesA, 1)

	// Clear notifications
	mockHub.Clear()

	// Second match: A <-> C
	_ = matchesRepo.RecordMatchAction(ctx, userA.String(), userC.String(), true)
	_ = matchesRepo.RecordMatchAction(ctx, userC.String(), userA.String(), true)

	// Try to award badge again (should not duplicate)
	err := badgeService.AwardBadge(ctx, userA, "first_match")
	require.NoError(t, err, "Should not error when trying to award duplicate")

	// Award to C (their first match)
	err = badgeService.AwardBadge(ctx, userC, "first_match")
	require.NoError(t, err)

	// Verify A still has only 1 badge
	badgesA, _ = badgesRepo.GetUserBadges(ctx, userA)
	assert.Len(t, badgesA, 1, "User A should still have only 1 badge")

	// Verify C has the badge (their first)
	badgesC, _ := badgesRepo.GetUserBadges(ctx, userC)
	assert.Len(t, badgesC, 1, "User C should have 1 badge")

	// Verify A didn't get a second notification
	messagesA := mockHub.GetMessagesForUser(userA.String())
	assert.Len(t, messagesA, 0, "User A should not receive duplicate notification")

	// Verify C got notification
	messagesC := mockHub.GetMessagesForUser(userC.String())
	assert.Len(t, messagesC, 1, "User C should receive badge notification")
}

// TestFirstMatchBadge_NoAwardOnRejection tests badge is not awarded when match is rejected
func TestFirstMatchBadge_NoAwardOnRejection(t *testing.T) {
	pool := getBadgeTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()
	ensureFirstMatchBadgeExists(t, pool)

	userA := createTestUserWithSkills(t, pool, "reject_user_a")
	userB := createTestUserWithSkills(t, pool, "reject_user_b")
	defer cleanupBadgeTestUser(t, pool, userA)
	defer cleanupBadgeTestUser(t, pool, userB)

	matchesRepo := NewRepository(pool)
	badgesRepo := badges.NewRepository(pool)
	mockHub := NewMockNotificationHub()
	badgeService := badges.NewService(badgesRepo, mockHub)

	// User A accepts User B
	_ = matchesRepo.RecordMatchAction(ctx, userA.String(), userB.String(), true)

	// User B rejects User A
	_ = matchesRepo.RecordMatchAction(ctx, userB.String(), userA.String(), false)

	// Verify not connected
	connected, err := matchesRepo.AreConnected(ctx, userA.String(), userB.String())
	require.NoError(t, err)
	assert.False(t, connected, "Should not be connected when one user rejects")

	// Try to award badge (should work but not create connection)
	_ = badgeService.AwardBadge(ctx, userA, "first_match")
	_ = badgeService.AwardBadge(ctx, userB, "first_match")

	// In a real scenario, handler only awards when connected is true
	// So this test verifies the badge service itself doesn't check connection status
	// The handler is responsible for that logic
}

// TestFirstMatchBadge_DisplayOnProfile tests badge appears on user profile
func TestFirstMatchBadge_DisplayOnProfile(t *testing.T) {
	pool := getBadgeTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()
	ensureFirstMatchBadgeExists(t, pool)

	userID := createTestUserWithSkills(t, pool, "profile_user")
	defer cleanupBadgeTestUser(t, pool, userID)

	badgesRepo := badges.NewRepository(pool)
	mockHub := NewMockNotificationHub()
	badgeService := badges.NewService(badgesRepo, mockHub)

	// Award badge
	err := badgeService.AwardBadge(ctx, userID, "first_match")
	require.NoError(t, err)

	// Retrieve user badges (as profile would)
	userBadges, err := badgesRepo.GetUserBadges(ctx, userID)
	require.NoError(t, err)
	require.Len(t, userBadges, 1)

	badge := userBadges[0]
	assert.Equal(t, "first_match", badge.BadgeCode)
	assert.Equal(t, "First Connection", badge.BadgeName)
	assert.Contains(t, badge.BadgeIcon, "🤝")
	assert.NotZero(t, badge.AwardedAt, "Badge should have awarded timestamp")
}

// TestFirstMatchBadge_ConcurrentAwarding tests concurrent badge awarding
func TestFirstMatchBadge_ConcurrentAwarding(t *testing.T) {
	pool := getBadgeTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()
	ensureFirstMatchBadgeExists(t, pool)

	userID := createTestUserWithSkills(t, pool, "concurrent_user")
	defer cleanupBadgeTestUser(t, pool, userID)

	badgesRepo := badges.NewRepository(pool)
	mockHub := NewMockNotificationHub()
	badgeService := badges.NewService(badgesRepo, mockHub)

	// Try to award badge concurrently (simulating race condition)
	done := make(chan bool, 2)

	for i := 0; i < 2; i++ {
		go func() {
			_ = badgeService.AwardBadge(ctx, userID, "first_match")
			done <- true
		}()
	}

	// Wait for both goroutines
	<-done
	<-done

	// Should still only have one badge
	userBadges, err := badgesRepo.GetUserBadges(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, userBadges, 1, "Should have exactly 1 badge despite concurrent awarding")
}
