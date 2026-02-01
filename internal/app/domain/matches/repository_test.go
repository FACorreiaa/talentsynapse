package matches

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestDBPool creates a connection pool to the test database
func getTestDBPool(t *testing.T) *pgxpool.Pool {
	// Try to get DSN from environment, fall back to default test DB
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

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("Skipping integration test: database ping failed: %v", err)
		return nil
	}

	return pool
}

// getBenchDBPool creates a connection pool for benchmarks
func getBenchDBPool(b *testing.B) *pgxpool.Pool {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5470/myapp?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		b.Skipf("Skipping benchmark: could not connect to database: %v", err)
		return nil
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		b.Skipf("Skipping benchmark: database not responding: %v", err)
		return nil
	}

	return pool
}

// testUser represents a user created for testing
type testUser struct {
	ID          string
	Email       string
	Username    string
	DisplayName string
}

// createTestUser creates a test user in the database
func createTestUser(t *testing.T, pool *pgxpool.Pool, username string) *testUser {
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("%s_%s@test.com", username, uniqueSuffix)

	query := `
		INSERT INTO users (email, username, hashed_password, display_name, is_active)
		VALUES ($1, $2, '$2a$10$test', $3, true)
		RETURNING id, email, username, display_name
	`

	u := &testUser{}
	err := pool.QueryRow(ctx, query, email, username, username).Scan(
		&u.ID, &u.Email, &u.Username, &u.DisplayName,
	)
	require.NoError(t, err)

	return u
}

// cleanupTestUser removes a test user and related data from the database
func cleanupTestUser(t *testing.T, pool *pgxpool.Pool, userID string) {
	ctx := context.Background()
	// Delete in correct order due to foreign keys
	_, _ = pool.Exec(ctx, "DELETE FROM match_history WHERE user_id_a = $1 OR user_id_b = $1", userID)
	_, _ = pool.Exec(ctx, "DELETE FROM user_skills WHERE user_id = $1", userID)
	_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
}

// recordMatchAction creates a match history entry
func recordMatchAction(t *testing.T, pool *pgxpool.Pool, userA, userB string, accepted bool) {
	ctx := context.Background()
	query := `
		INSERT INTO match_history (user_id_a, user_id_b, algorithm_used, match_score, interaction_initiated)
		VALUES ($1, $2, 'cosine_similarity', 0.5, $3)
		ON CONFLICT (user_id_a, user_id_b, created_at) DO NOTHING
	`
	_, err := pool.Exec(ctx, query, userA, userB, accepted)
	require.NoError(t, err)
}

// TestGetMutualMatches_Integration tests GetMutualMatches with real database
func TestGetMutualMatches_Integration(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewRepository(pool)
	ctx := context.Background()

	// Create test users
	userA := createTestUser(t, pool, "test_user_a")
	userB := createTestUser(t, pool, "test_user_b")
	userC := createTestUser(t, pool, "test_user_c")
	userD := createTestUser(t, pool, "test_user_d")

	defer func() {
		cleanupTestUser(t, pool, userA.ID)
		cleanupTestUser(t, pool, userB.ID)
		cleanupTestUser(t, pool, userC.ID)
		cleanupTestUser(t, pool, userD.ID)
	}()

	t.Run("returns empty when no mutual matches", func(t *testing.T) {
		matches, err := repo.GetMutualMatches(ctx, userA.ID, 10)
		require.NoError(t, err)
		assert.Empty(t, matches)
	})

	t.Run("does not return one-sided matches - A accepted B but B did not accept A", func(t *testing.T) {
		// A accepts B
		recordMatchAction(t, pool, userA.ID, userB.ID, true)

		matches, err := repo.GetMutualMatches(ctx, userA.ID, 10)
		require.NoError(t, err)
		assert.Empty(t, matches, "should not return B since B has not accepted A")
	})

	t.Run("returns mutual match when both users accepted each other", func(t *testing.T) {
		// B accepts A (A already accepted B in previous test)
		recordMatchAction(t, pool, userB.ID, userA.ID, true)

		// Check from A's perspective
		matchesA, err := repo.GetMutualMatches(ctx, userA.ID, 10)
		require.NoError(t, err)
		assert.Len(t, matchesA, 1, "A should see B as mutual match")
		assert.Equal(t, userB.ID, matchesA[0].UserID)
		assert.Equal(t, userB.DisplayName, matchesA[0].DisplayName)
		assert.Equal(t, userB.Username, matchesA[0].Username)

		// Check from B's perspective
		matchesB, err := repo.GetMutualMatches(ctx, userB.ID, 10)
		require.NoError(t, err)
		assert.Len(t, matchesB, 1, "B should see A as mutual match")
		assert.Equal(t, userA.ID, matchesB[0].UserID)
	})

	t.Run("does not return rejected matches", func(t *testing.T) {
		// C accepts A, A rejects C
		recordMatchAction(t, pool, userC.ID, userA.ID, true)
		recordMatchAction(t, pool, userA.ID, userC.ID, false)

		matches, err := repo.GetMutualMatches(ctx, userA.ID, 10)
		require.NoError(t, err)
		// Should only have B, not C
		assert.Len(t, matches, 1)
		assert.Equal(t, userB.ID, matches[0].UserID)
	})

	t.Run("returns multiple mutual matches", func(t *testing.T) {
		// Create mutual match between A and D
		recordMatchAction(t, pool, userA.ID, userD.ID, true)
		recordMatchAction(t, pool, userD.ID, userA.ID, true)

		matches, err := repo.GetMutualMatches(ctx, userA.ID, 10)
		require.NoError(t, err)
		assert.Len(t, matches, 2, "A should have 2 mutual matches: B and D")

		// Verify both are present
		userIDs := []string{matches[0].UserID, matches[1].UserID}
		assert.Contains(t, userIDs, userB.ID)
		assert.Contains(t, userIDs, userD.ID)
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		matches, err := repo.GetMutualMatches(ctx, userA.ID, 1)
		require.NoError(t, err)
		assert.Len(t, matches, 1, "should respect limit of 1")
	})

	t.Run("does not include inactive users", func(t *testing.T) {
		// Deactivate userB
		_, err := pool.Exec(ctx, "UPDATE users SET is_active = false WHERE id = $1", userB.ID)
		require.NoError(t, err)

		matches, err := repo.GetMutualMatches(ctx, userA.ID, 10)
		require.NoError(t, err)
		// Should only have D now
		assert.Len(t, matches, 1)
		assert.Equal(t, userD.ID, matches[0].UserID)

		// Reactivate for cleanup
		_, _ = pool.Exec(ctx, "UPDATE users SET is_active = true WHERE id = $1", userB.ID)
	})
}

// TestAreConnected_Integration tests AreConnected with real database
func TestAreConnected_Integration(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewRepository(pool)
	ctx := context.Background()

	// Create test users
	userA := createTestUser(t, pool, "test_conn_a")
	userB := createTestUser(t, pool, "test_conn_b")
	userC := createTestUser(t, pool, "test_conn_c")

	defer func() {
		cleanupTestUser(t, pool, userA.ID)
		cleanupTestUser(t, pool, userB.ID)
		cleanupTestUser(t, pool, userC.ID)
	}()

	t.Run("returns false when no connection exists", func(t *testing.T) {
		connected, err := repo.AreConnected(ctx, userA.ID, userB.ID)
		require.NoError(t, err)
		assert.False(t, connected)
	})

	t.Run("returns false when only one side accepted", func(t *testing.T) {
		// A accepts B
		recordMatchAction(t, pool, userA.ID, userB.ID, true)

		// Check both directions
		connectedAB, err := repo.AreConnected(ctx, userA.ID, userB.ID)
		require.NoError(t, err)
		assert.False(t, connectedAB, "should be false when only A accepted B")

		connectedBA, err := repo.AreConnected(ctx, userB.ID, userA.ID)
		require.NoError(t, err)
		assert.False(t, connectedBA, "should be false regardless of parameter order")
	})

	t.Run("returns false when one side rejected", func(t *testing.T) {
		// B rejects A
		recordMatchAction(t, pool, userB.ID, userA.ID, false)

		connected, err := repo.AreConnected(ctx, userA.ID, userB.ID)
		require.NoError(t, err)
		assert.False(t, connected, "should be false when one side rejected")
	})

	t.Run("returns true when both accepted - update B to accept A", func(t *testing.T) {
		// Delete the rejection and add acceptance
		_, err := pool.Exec(ctx, "DELETE FROM match_history WHERE user_id_a = $1 AND user_id_b = $2", userB.ID, userA.ID)
		require.NoError(t, err)
		recordMatchAction(t, pool, userB.ID, userA.ID, true)

		// Now both have accepted
		connected, err := repo.AreConnected(ctx, userA.ID, userB.ID)
		require.NoError(t, err)
		assert.True(t, connected, "should be true when both accepted")
	})

	t.Run("connection check is symmetric", func(t *testing.T) {
		// Check both parameter orders
		connectedAB, err := repo.AreConnected(ctx, userA.ID, userB.ID)
		require.NoError(t, err)

		connectedBA, err := repo.AreConnected(ctx, userB.ID, userA.ID)
		require.NoError(t, err)

		assert.Equal(t, connectedAB, connectedBA, "connection check should be symmetric")
		assert.True(t, connectedAB, "both should be true")
	})

	t.Run("returns false for users who never interacted", func(t *testing.T) {
		connected, err := repo.AreConnected(ctx, userA.ID, userC.ID)
		require.NoError(t, err)
		assert.False(t, connected, "should be false for users with no interaction")
	})

	t.Run("returns false for same user ID", func(t *testing.T) {
		connected, err := repo.AreConnected(ctx, userA.ID, userA.ID)
		require.NoError(t, err)
		assert.False(t, connected, "user should not be connected to themselves")
	})
}

// TestGetConnectionCount_Integration tests GetConnectionCount with real database
func TestGetConnectionCount_Integration(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewRepository(pool)
	ctx := context.Background()

	// Create test users
	userA := createTestUser(t, pool, "test_count_a")
	userB := createTestUser(t, pool, "test_count_b")
	userC := createTestUser(t, pool, "test_count_c")
	userD := createTestUser(t, pool, "test_count_d")
	userE := createTestUser(t, pool, "test_count_e")

	defer func() {
		cleanupTestUser(t, pool, userA.ID)
		cleanupTestUser(t, pool, userB.ID)
		cleanupTestUser(t, pool, userC.ID)
		cleanupTestUser(t, pool, userD.ID)
		cleanupTestUser(t, pool, userE.ID)
	}()

	t.Run("returns 0 when user has no connections", func(t *testing.T) {
		count, err := repo.GetConnectionCount(ctx, userA.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("returns 0 when only one-sided matches exist", func(t *testing.T) {
		// A accepts B and C
		recordMatchAction(t, pool, userA.ID, userB.ID, true)
		recordMatchAction(t, pool, userA.ID, userC.ID, true)

		count, err := repo.GetConnectionCount(ctx, userA.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "one-sided matches should not count")
	})

	t.Run("returns 1 when one mutual connection exists", func(t *testing.T) {
		// B accepts A back
		recordMatchAction(t, pool, userB.ID, userA.ID, true)

		count, err := repo.GetConnectionCount(ctx, userA.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "should count one mutual connection")

		// Verify from B's perspective too
		countB, err := repo.GetConnectionCount(ctx, userB.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, countB, "B should also see 1 connection")
	})

	t.Run("returns correct count with multiple mutual connections", func(t *testing.T) {
		// C accepts A back
		recordMatchAction(t, pool, userC.ID, userA.ID, true)
		// Create mutual connection with D
		recordMatchAction(t, pool, userA.ID, userD.ID, true)
		recordMatchAction(t, pool, userD.ID, userA.ID, true)

		count, err := repo.GetConnectionCount(ctx, userA.ID)
		require.NoError(t, err)
		assert.Equal(t, 3, count, "A should have 3 mutual connections: B, C, D")
	})

	t.Run("does not count rejected matches", func(t *testing.T) {
		// A accepts E but E rejects A
		recordMatchAction(t, pool, userA.ID, userE.ID, true)
		recordMatchAction(t, pool, userE.ID, userA.ID, false)

		count, err := repo.GetConnectionCount(ctx, userA.ID)
		require.NoError(t, err)
		assert.Equal(t, 3, count, "rejected match should not be counted")
	})

	t.Run("does not count inactive users", func(t *testing.T) {
		// Deactivate userB
		_, err := pool.Exec(ctx, "UPDATE users SET is_active = false WHERE id = $1", userB.ID)
		require.NoError(t, err)

		count, err := repo.GetConnectionCount(ctx, userA.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, count, "inactive user should not be counted, A should have C and D")

		// Reactivate for cleanup
		_, _ = pool.Exec(ctx, "UPDATE users SET is_active = true WHERE id = $1", userB.ID)
	})

	t.Run("count is consistent from both users' perspectives", func(t *testing.T) {
		// Check that A-B connection is counted the same way from both sides
		// First reactivate B
		_, err := pool.Exec(ctx, "UPDATE users SET is_active = true WHERE id = $1", userB.ID)
		require.NoError(t, err)

		// Get A's connections that include B
		mutualMatches, err := repo.GetMutualMatches(ctx, userA.ID, 10)
		require.NoError(t, err)

		hasB := false
		for _, m := range mutualMatches {
			if m.UserID == userB.ID {
				hasB = true
				break
			}
		}
		assert.True(t, hasB, "B should be in A's mutual matches")

		// B should see A in their mutual matches too
		mutualMatchesB, err := repo.GetMutualMatches(ctx, userB.ID, 10)
		require.NoError(t, err)

		hasA := false
		for _, m := range mutualMatchesB {
			if m.UserID == userA.ID {
				hasA = true
				break
			}
		}
		assert.True(t, hasA, "A should be in B's mutual matches")
	})
}

// TestHasConnections_Integration tests HasConnections helper function
func TestHasConnections_Integration(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewRepository(pool)
	ctx := context.Background()

	userA := createTestUser(t, pool, "test_has_a")
	userB := createTestUser(t, pool, "test_has_b")

	defer func() {
		cleanupTestUser(t, pool, userA.ID)
		cleanupTestUser(t, pool, userB.ID)
	}()

	t.Run("returns false when no connections", func(t *testing.T) {
		has, err := repo.HasConnections(ctx, userA.ID)
		require.NoError(t, err)
		assert.False(t, has)
	})

	t.Run("returns true when at least one connection exists", func(t *testing.T) {
		// Create mutual connection
		recordMatchAction(t, pool, userA.ID, userB.ID, true)
		recordMatchAction(t, pool, userB.ID, userA.ID, true)

		has, err := repo.HasConnections(ctx, userA.ID)
		require.NoError(t, err)
		assert.True(t, has)
	})
}

// TestMatchActionRecording_Integration tests RecordMatchAction
func TestMatchActionRecording_Integration(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewRepository(pool)
	ctx := context.Background()

	// Create test users with skills for match score calculation
	userA := createTestUser(t, pool, "test_record_a")
	userB := createTestUser(t, pool, "test_record_b")

	defer func() {
		cleanupTestUser(t, pool, userA.ID)
		cleanupTestUser(t, pool, userB.ID)
	}()

	t.Run("records acceptance action", func(t *testing.T) {
		err := repo.RecordMatchAction(ctx, userA.ID, userB.ID, true)
		require.NoError(t, err)

		// Verify it was recorded
		var exists bool
		query := `SELECT EXISTS(SELECT 1 FROM match_history WHERE user_id_a = $1 AND user_id_b = $2 AND interaction_initiated = true)`
		err = pool.QueryRow(ctx, query, userA.ID, userB.ID).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("records rejection action", func(t *testing.T) {
		err := repo.RecordMatchAction(ctx, userB.ID, userA.ID, false)
		require.NoError(t, err)

		// Verify it was recorded
		var exists bool
		query := `SELECT EXISTS(SELECT 1 FROM match_history WHERE user_id_a = $1 AND user_id_b = $2 AND interaction_initiated = false)`
		err = pool.QueryRow(ctx, query, userB.ID, userA.ID).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists)
	})
}

// TestEdgeCases_Integration tests various edge cases
func TestEdgeCases_Integration(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewRepository(pool)
	ctx := context.Background()

	t.Run("GetMutualMatches with empty database", func(t *testing.T) {
		// Use a UUID that doesn't exist
		matches, err := repo.GetMutualMatches(ctx, "00000000-0000-0000-0000-000000000000", 10)
		require.NoError(t, err)
		assert.Empty(t, matches)
	})

	t.Run("GetConnectionCount with empty database", func(t *testing.T) {
		count, err := repo.GetConnectionCount(ctx, "00000000-0000-0000-0000-000000000000")
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("AreConnected with non-existent users", func(t *testing.T) {
		connected, err := repo.AreConnected(ctx,
			"00000000-0000-0000-0000-000000000000",
			"00000000-0000-0000-0000-000000000001")
		require.NoError(t, err)
		assert.False(t, connected)
	})

	t.Run("GetMutualMatches with zero limit", func(t *testing.T) {
		userA := createTestUser(t, pool, "test_limit_zero")
		defer cleanupTestUser(t, pool, userA.ID)

		matches, err := repo.GetMutualMatches(ctx, userA.ID, 0)
		require.NoError(t, err)
		assert.Empty(t, matches)
	})

	t.Run("GetMutualMatches with negative limit returns error", func(t *testing.T) {
		userA := createTestUser(t, pool, "test_limit_neg")
		defer cleanupTestUser(t, pool, userA.ID)

		matches, err := repo.GetMutualMatches(ctx, userA.ID, -1)
		assert.Error(t, err, "negative limit should return an error")
		assert.Empty(t, matches)
	})
}
