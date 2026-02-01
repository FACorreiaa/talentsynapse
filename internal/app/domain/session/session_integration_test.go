package session

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/points"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

// createTestUser creates a test user
func createTestUser(t *testing.T, pool *pgxpool.Pool, email, username string) string {
	ctx := context.Background()

	query := `
		INSERT INTO users (email, username, hashed_password, display_name, is_active)
		VALUES ($1, $2, '$2a$10$test', $3, true)
		RETURNING id
	`

	var userID string
	err := pool.QueryRow(ctx, query, email, username, username).Scan(&userID)
	require.NoError(t, err)

	return userID
}

// createMutualConnection creates a mutual match between two users
func createMutualConnection(t *testing.T, pool *pgxpool.Pool, user1ID, user2ID string) {
	ctx := context.Background()

	// User1 accepts User2
	_, err := pool.Exec(ctx, `
		INSERT INTO match_history (user_id_a, user_id_b, algorithm_used, match_score, interaction_initiated)
		VALUES ($1::uuid, $2::uuid, 'test', 0.8, true)
	`, user1ID, user2ID)
	require.NoError(t, err)

	// User2 accepts User1
	_, err = pool.Exec(ctx, `
		INSERT INTO match_history (user_id_a, user_id_b, algorithm_used, match_score, interaction_initiated)
		VALUES ($1::uuid, $2::uuid, 'test', 0.8, true)
	`, user2ID, user1ID)
	require.NoError(t, err)
}

// cleanupTestData removes test data
func cleanupTestData(t *testing.T, pool *pgxpool.Pool, userIDs ...string) {
	ctx := context.Background()
	for _, userID := range userIDs {
		_, _ = pool.Exec(ctx, "DELETE FROM sessions WHERE initiator_id = $1 OR partner_id = $1", userID)
		_, _ = pool.Exec(ctx, "DELETE FROM match_history WHERE user_id_a = $1 OR user_id_b = $1", userID)
		_, _ = pool.Exec(ctx, "DELETE FROM user_stats WHERE user_id = $1", userID)
		_, _ = pool.Exec(ctx, "DELETE FROM points_history WHERE user_id = $1", userID)
		_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	}
}

// TestSessionLifecycle_Integration tests the complete session lifecycle
func TestSessionLifecycle_Integration(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewRepository(pool)
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create two connected users
	user1ID := createTestUser(t, pool,
		fmt.Sprintf("session1_%s@test.com", uniqueSuffix),
		fmt.Sprintf("session1_%s", uniqueSuffix),
	)
	user2ID := createTestUser(t, pool,
		fmt.Sprintf("session2_%s@test.com", uniqueSuffix),
		fmt.Sprintf("session2_%s", uniqueSuffix),
	)
	defer cleanupTestData(t, pool, user1ID, user2ID)

	// Create mutual connection
	createMutualConnection(t, pool, user1ID, user2ID)

	var sessionID string

	t.Run("create session request", func(t *testing.T) {
		startTime := time.Now().Add(24 * time.Hour)
		err := repo.CreateRequest(ctx, user1ID, user2ID, startTime)
		require.NoError(t, err)

		// Verify session was created
		sessions, err := repo.GetUserSessions(ctx, user1ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(sessions), 1, "Should have at least one session")

		// Find the created session
		for _, s := range sessions {
			if s.InitiatorID == user1ID && s.PartnerID == user2ID {
				sessionID = s.ID
				assert.Equal(t, "pending", s.Status, "New session should be pending")
				break
			}
		}
		assert.NotEmpty(t, sessionID, "Session ID should be set")
	})

	t.Run("both users see the session", func(t *testing.T) {
		// User1 sees the session
		sessions1, err := repo.GetUserSessions(ctx, user1ID)
		require.NoError(t, err)

		found1 := false
		for _, s := range sessions1 {
			if s.ID == sessionID {
				found1 = true
				assert.Equal(t, user1ID, s.InitiatorID)
				assert.Equal(t, user2ID, s.PartnerID)
				break
			}
		}
		assert.True(t, found1, "User1 should see the session")

		// User2 sees the session
		sessions2, err := repo.GetUserSessions(ctx, user2ID)
		require.NoError(t, err)

		found2 := false
		for _, s := range sessions2 {
			if s.ID == sessionID {
				found2 = true
				break
			}
		}
		assert.True(t, found2, "User2 should see the session")
	})

	t.Run("mark session as completed", func(t *testing.T) {
		err := repo.MarkCompleted(ctx, sessionID, user1ID)
		require.NoError(t, err)

		// Verify session status updated
		session, err := repo.GetSessionByID(ctx, sessionID)
		require.NoError(t, err)
		assert.Equal(t, "completed", session.Status, "Session should be completed")
	})

	t.Run("count completed sessions", func(t *testing.T) {
		count, err := repo.CountCompletedSessions(ctx, user1ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1, "Should have at least one completed session")
	})

	t.Run("only participants can complete session", func(t *testing.T) {
		// Create another session
		startTime := time.Now().Add(48 * time.Hour)
		err := repo.CreateRequest(ctx, user1ID, user2ID, startTime)
		require.NoError(t, err)

		// Create a third user (non-participant)
		user3ID := createTestUser(t, pool,
			fmt.Sprintf("nonparticipant_%s@test.com", uniqueSuffix),
			fmt.Sprintf("nonparticipant_%s", uniqueSuffix),
		)
		defer cleanupTestData(t, pool, user3ID)

		// Get the new session ID
		sessions, err := repo.GetUserSessions(ctx, user1ID)
		require.NoError(t, err)
		var newSessionID string
		for _, s := range sessions {
			if s.Status == "pending" && s.InitiatorID == user1ID && s.PartnerID == user2ID {
				newSessionID = s.ID
				break
			}
		}
		require.NotEmpty(t, newSessionID, "Should have a new pending session")

		// Try to complete with non-participant
		err = repo.MarkCompleted(ctx, newSessionID, user3ID)
		require.NoError(t, err) // Query succeeds but doesn't update

		// Verify status unchanged
		session, err := repo.GetSessionByID(ctx, newSessionID)
		require.NoError(t, err)
		assert.Equal(t, "pending", session.Status, "Status should remain pending")
	})
}

// TestSessionLifecycle_WithPoints tests session completion with points
func TestSessionLifecycle_WithPoints(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	sessionRepo := NewRepository(pool)
	pointsRepo := points.NewRepository(pool)
	pointsService := points.NewService(pointsRepo, nil)
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create two users
	user1ID := createTestUser(t, pool,
		fmt.Sprintf("points1_%s@test.com", uniqueSuffix),
		fmt.Sprintf("points1_%s", uniqueSuffix),
	)
	user2ID := createTestUser(t, pool,
		fmt.Sprintf("points2_%s@test.com", uniqueSuffix),
		fmt.Sprintf("points2_%s", uniqueSuffix),
	)
	defer cleanupTestData(t, pool, user1ID, user2ID)

	// Create mutual connection
	createMutualConnection(t, pool, user1ID, user2ID)

	// Ensure user stats exist
	err := pointsService.EnsureUserStats(ctx, uuid.MustParse(user1ID))
	require.NoError(t, err)
	err = pointsService.EnsureUserStats(ctx, uuid.MustParse(user2ID))
	require.NoError(t, err)

	t.Run("complete session and award points", func(t *testing.T) {
		// Create session
		startTime := time.Now().Add(1 * time.Hour)
		err := sessionRepo.CreateRequest(ctx, user1ID, user2ID, startTime)
		require.NoError(t, err)

		// Get session ID
		sessions, err := sessionRepo.GetUserSessions(ctx, user1ID)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(sessions), 1)
		sessionID := sessions[0].ID

		// Get initial points for both users
		initialPoints1, err := pointsService.GetUserPoints(ctx, uuid.MustParse(user1ID))
		require.NoError(t, err)
		initialPoints2, err := pointsService.GetUserPoints(ctx, uuid.MustParse(user2ID))
		require.NoError(t, err)

		// Mark session as completed
		err = sessionRepo.MarkCompleted(ctx, sessionID, user1ID)
		require.NoError(t, err)

		// Award points to both participants
		upgrade1, err := pointsService.AwardSessionPoints(ctx, uuid.MustParse(user1ID))
		require.NoError(t, err)

		upgrade2, err := pointsService.AwardSessionPoints(ctx, uuid.MustParse(user2ID))
		require.NoError(t, err)

		// Verify points awarded
		updatedPoints1, err := pointsService.GetUserPoints(ctx, uuid.MustParse(user1ID))
		require.NoError(t, err)
		updatedPoints2, err := pointsService.GetUserPoints(ctx, uuid.MustParse(user2ID))
		require.NoError(t, err)

		expectedIncrease := points.PointsReward[points.ActionSessionCompleted]
		assert.Equal(t, initialPoints1.Points+expectedIncrease, updatedPoints1.Points, "User1 should receive session points")
		assert.Equal(t, initialPoints2.Points+expectedIncrease, updatedPoints2.Points, "User2 should receive session points")

		// Log any tier upgrades
		if upgrade1 != nil {
			t.Logf("User1 upgraded from %s to %s", upgrade1.OldTier, upgrade1.NewTier)
		}
		if upgrade2 != nil {
			t.Logf("User2 upgraded from %s to %s", upgrade2.OldTier, upgrade2.NewTier)
		}
	})
}

// TestSessionLifecycle_MultipleSessions tests multiple sessions between users
func TestSessionLifecycle_MultipleSession(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewRepository(pool)
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create two users
	user1ID := createTestUser(t, pool,
		fmt.Sprintf("multi1_%s@test.com", uniqueSuffix),
		fmt.Sprintf("multi1_%s", uniqueSuffix),
	)
	user2ID := createTestUser(t, pool,
		fmt.Sprintf("multi2_%s@test.com", uniqueSuffix),
		fmt.Sprintf("multi2_%s", uniqueSuffix),
	)
	defer cleanupTestData(t, pool, user1ID, user2ID)

	// Create mutual connection
	createMutualConnection(t, pool, user1ID, user2ID)

	t.Run("create and complete multiple sessions", func(t *testing.T) {
		sessionCount := 3

		for i := 0; i < sessionCount; i++ {
			startTime := time.Now().Add(time.Duration(i*24) * time.Hour)
			err := repo.CreateRequest(ctx, user1ID, user2ID, startTime)
			require.NoError(t, err)
		}

		// Get all sessions
		sessions, err := repo.GetUserSessions(ctx, user1ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(sessions), sessionCount, "Should have at least %d sessions", sessionCount)

		// Complete all sessions
		for _, s := range sessions {
			if s.Status == "pending" {
				err := repo.MarkCompleted(ctx, s.ID, user1ID)
				require.NoError(t, err)
			}
		}

		// Count completed sessions
		count, err := repo.CountCompletedSessions(ctx, user1ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, sessionCount, "Should have completed %d sessions", sessionCount)
	})
}

// TestSessionLifecycle_OnlyConnectedUsers tests sessions only work between connected users
func TestSessionLifecycle_OnlyConnectedUsers(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewRepository(pool)
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create three users
	user1ID := createTestUser(t, pool,
		fmt.Sprintf("conn1_%s@test.com", uniqueSuffix),
		fmt.Sprintf("conn1_%s", uniqueSuffix),
	)
	user2ID := createTestUser(t, pool,
		fmt.Sprintf("conn2_%s@test.com", uniqueSuffix),
		fmt.Sprintf("conn2_%s", uniqueSuffix),
	)
	user3ID := createTestUser(t, pool,
		fmt.Sprintf("conn3_%s@test.com", uniqueSuffix),
		fmt.Sprintf("conn3_%s", uniqueSuffix),
	)
	defer cleanupTestData(t, pool, user1ID, user2ID, user3ID)

	// User1 and User2 are connected
	createMutualConnection(t, pool, user1ID, user2ID)

	// User1 and User3 are NOT connected

	t.Run("session visible only between connected users", func(t *testing.T) {
		// Create session between connected users (User1 and User2)
		startTime := time.Now().Add(1 * time.Hour)
		err := repo.CreateRequest(ctx, user1ID, user2ID, startTime)
		require.NoError(t, err)

		// User1 should see it
		sessions1, err := repo.GetUserSessions(ctx, user1ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(sessions1), 1, "User1 should see sessions")

		// User2 should see it
		sessions2, err := repo.GetUserSessions(ctx, user2ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(sessions2), 1, "User2 should see sessions")

		// User3 should NOT see it (not connected)
		sessions3, err := repo.GetUserSessions(ctx, user3ID)
		require.NoError(t, err)
		assert.Equal(t, 0, len(sessions3), "User3 should not see sessions from unconnected users")
	})

	t.Run("create session between non-connected users", func(t *testing.T) {
		// Technically allowed to create, but won't show in GetUserSessions due to mutual connection check
		startTime := time.Now().Add(2 * time.Hour)
		err := repo.CreateRequest(ctx, user1ID, user3ID, startTime)
		require.NoError(t, err)

		// User1 should NOT see it in GetUserSessions (requires mutual connection)
		sessions1, err := repo.GetUserSessions(ctx, user1ID)
		require.NoError(t, err)

		foundNonConnectedSession := false
		for _, s := range sessions1 {
			if s.PartnerID == user3ID {
				foundNonConnectedSession = true
				break
			}
		}
		assert.False(t, foundNonConnectedSession, "Session with non-connected user should not appear")
	})
}

// TestSessionLifecycle_GetSessionByID tests retrieving individual sessions
func TestSessionLifecycle_GetSessionByID(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewRepository(pool)
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	user1ID := createTestUser(t, pool,
		fmt.Sprintf("getid1_%s@test.com", uniqueSuffix),
		fmt.Sprintf("getid1_%s", uniqueSuffix),
	)
	user2ID := createTestUser(t, pool,
		fmt.Sprintf("getid2_%s@test.com", uniqueSuffix),
		fmt.Sprintf("getid2_%s", uniqueSuffix),
	)
	defer cleanupTestData(t, pool, user1ID, user2ID)

	createMutualConnection(t, pool, user1ID, user2ID)

	t.Run("get session by ID", func(t *testing.T) {
		startTime := time.Now().Add(1 * time.Hour)
		err := repo.CreateRequest(ctx, user1ID, user2ID, startTime)
		require.NoError(t, err)

		// Get all sessions to find the ID
		sessions, err := repo.GetUserSessions(ctx, user1ID)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(sessions), 1)
		sessionID := sessions[0].ID

		// Get session by ID
		session, err := repo.GetSessionByID(ctx, sessionID)
		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, sessionID, session.ID)
		assert.Equal(t, user1ID, session.InitiatorID)
		assert.Equal(t, user2ID, session.PartnerID)
	})

	t.Run("get non-existent session returns error", func(t *testing.T) {
		fakeID := uuid.New().String()
		session, err := repo.GetSessionByID(ctx, fakeID)
		assert.Error(t, err, "Should error for non-existent session")
		assert.Nil(t, session)
	})
}

// TestSessionLifecycle_CountCompletedSessions tests session counting
func TestSessionLifecycle_CountCompletedSessions(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewRepository(pool)
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	user1ID := createTestUser(t, pool,
		fmt.Sprintf("count1_%s@test.com", uniqueSuffix),
		fmt.Sprintf("count1_%s", uniqueSuffix),
	)
	user2ID := createTestUser(t, pool,
		fmt.Sprintf("count2_%s@test.com", uniqueSuffix),
		fmt.Sprintf("count2_%s", uniqueSuffix),
	)
	defer cleanupTestData(t, pool, user1ID, user2ID)

	createMutualConnection(t, pool, user1ID, user2ID)

	t.Run("count starts at zero", func(t *testing.T) {
		count, err := repo.CountCompletedSessions(ctx, user1ID)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Should start with zero completed sessions")
	})

	t.Run("count increases after completion", func(t *testing.T) {
		// Create and complete 3 sessions
		for i := 0; i < 3; i++ {
			startTime := time.Now().Add(time.Duration(i) * time.Hour)
			err := repo.CreateRequest(ctx, user1ID, user2ID, startTime)
			require.NoError(t, err)
		}

		// Get sessions and complete them
		sessions, err := repo.GetUserSessions(ctx, user1ID)
		require.NoError(t, err)

		for _, s := range sessions {
			if s.Status == "pending" {
				err := repo.MarkCompleted(ctx, s.ID, user1ID)
				require.NoError(t, err)
			}
		}

		// Count should be 3
		count, err := repo.CountCompletedSessions(ctx, user1ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 3, "Should have at least 3 completed sessions")
	})
}
