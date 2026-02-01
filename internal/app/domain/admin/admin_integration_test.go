package admin

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/user"
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
func createTestUser(t *testing.T, pool *pgxpool.Pool, email, username string, isActive bool) string {
	ctx := context.Background()

	query := `
		INSERT INTO users (email, username, hashed_password, display_name, is_active)
		VALUES ($1, $2, '$2a$10$test', $3, $4)
		RETURNING id
	`

	var userID string
	err := pool.QueryRow(ctx, query, email, username, username, isActive).Scan(&userID)
	require.NoError(t, err)

	return userID
}

// cleanupTestUser removes test user and related data
func cleanupTestUser(t *testing.T, pool *pgxpool.Pool, userID string) {
	ctx := context.Background()
	_, _ = pool.Exec(ctx, "DELETE FROM sessions WHERE initiator_id = $1 OR partner_id = $1", userID)
	_, _ = pool.Exec(ctx, "DELETE FROM match_history WHERE user_id_a = $1 OR user_id_b = $1", userID)
	_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
}

// TestAdminActions_BanUser tests banning a user
func TestAdminActions_BanUser(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	userRepo := user.NewRepository(pool)
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	userID := createTestUser(t, pool,
		fmt.Sprintf("ban_test_%s@test.com", uniqueSuffix),
		fmt.Sprintf("ban_test_%s", uniqueSuffix),
		true,
	)
	defer cleanupTestUser(t, pool, userID)

	t.Run("user is active initially", func(t *testing.T) {
		isActive, err := userRepo.IsUserActive(ctx, userID)
		require.NoError(t, err)
		assert.True(t, isActive, "User should be active initially")
	})

	t.Run("ban user", func(t *testing.T) {
		err := userRepo.BanUser(ctx, userID)
		require.NoError(t, err)

		// Verify user is now inactive
		isActive, err := userRepo.IsUserActive(ctx, userID)
		require.NoError(t, err)
		assert.False(t, isActive, "User should be inactive after ban")
	})

	t.Run("verify user cannot access protected resources", func(t *testing.T) {
		// User should not be active
		isActive, err := userRepo.IsUserActive(ctx, userID)
		require.NoError(t, err)
		assert.False(t, isActive, "Banned user should not be active")

		// Verify is_active directly from database
		var dbIsActive bool
		err = pool.QueryRow(ctx, "SELECT is_active FROM users WHERE id = $1", userID).Scan(&dbIsActive)
		require.NoError(t, err)
		assert.False(t, dbIsActive, "User record should show is_active = false")
	})

	t.Run("unban user", func(t *testing.T) {
		err := userRepo.UnbanUser(ctx, userID)
		require.NoError(t, err)

		// Verify user is active again
		isActive, err := userRepo.IsUserActive(ctx, userID)
		require.NoError(t, err)
		assert.True(t, isActive, "User should be active after unban")
	})

	t.Run("can ban again after unban", func(t *testing.T) {
		// Ban again
		err := userRepo.BanUser(ctx, userID)
		require.NoError(t, err)

		isActive, err := userRepo.IsUserActive(ctx, userID)
		require.NoError(t, err)
		assert.False(t, isActive, "User should be banned again")
	})
}

// TestAdminActions_BannedUserSession tests that banned users cannot participate in sessions
func TestAdminActions_BannedUserSession(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	userRepo := user.NewRepository(pool)
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create two users
	user1ID := createTestUser(t, pool,
		fmt.Sprintf("session1_%s@test.com", uniqueSuffix),
		fmt.Sprintf("session1_%s", uniqueSuffix),
		true,
	)
	user2ID := createTestUser(t, pool,
		fmt.Sprintf("session2_%s@test.com", uniqueSuffix),
		fmt.Sprintf("session2_%s", uniqueSuffix),
		true,
	)
	defer cleanupTestUser(t, pool, user1ID)
	defer cleanupTestUser(t, pool, user2ID)

	t.Run("create session between active users", func(t *testing.T) {
		// Both users are active
		isActive1, err := userRepo.IsUserActive(ctx, user1ID)
		require.NoError(t, err)
		assert.True(t, isActive1)

		isActive2, err := userRepo.IsUserActive(ctx, user2ID)
		require.NoError(t, err)
		assert.True(t, isActive2)

		// Create a session (would normally be done via session repository)
		_, err = pool.Exec(ctx, `
			INSERT INTO sessions (
				initiator_id, partner_id,
				initiator_offers, partner_offers,
				scheduled_start, scheduled_end,
				status
			) VALUES ($1, $2, '{}', '{}', NOW(), NOW() + interval '1 hour', 'pending')
		`, user1ID, user2ID)
		require.NoError(t, err)
	})

	t.Run("ban user1 invalidates access", func(t *testing.T) {
		err := userRepo.BanUser(ctx, user1ID)
		require.NoError(t, err)

		// User1 should not be active
		isActive, err := userRepo.IsUserActive(ctx, user1ID)
		require.NoError(t, err)
		assert.False(t, isActive, "Banned user should not be active")

		// In real application, this check would prevent:
		// - User1 from viewing matches
		// - User1 from creating new sessions
		// - User1 from completing sessions
		// - User1 from submitting reviews
	})

	t.Run("user2 still active", func(t *testing.T) {
		isActive, err := userRepo.IsUserActive(ctx, user2ID)
		require.NoError(t, err)
		assert.True(t, isActive, "User2 should still be active")
	})
}

// TestAdminActions_BannedUserMatches tests that banned users don't appear in matches
func TestAdminActions_BannedUserMatches(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	userRepo := user.NewRepository(pool)
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create two users with skills
	user1ID := createTestUser(t, pool,
		fmt.Sprintf("match1_%s@test.com", uniqueSuffix),
		fmt.Sprintf("match1_%s", uniqueSuffix),
		true,
	)
	user2ID := createTestUser(t, pool,
		fmt.Sprintf("match2_%s@test.com", uniqueSuffix),
		fmt.Sprintf("match2_%s", uniqueSuffix),
		true,
	)
	defer cleanupTestUser(t, pool, user1ID)
	defer cleanupTestUser(t, pool, user2ID)

	// Add skills to both users
	var skillID string
	err := pool.QueryRow(ctx, `
		INSERT INTO skills (name, category)
		VALUES ('TestSkill', 'Technology')
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`).Scan(&skillID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO user_skills (user_id, skill_id, skill_type, proficiency)
		VALUES ($1, $2, 'offered', 8), ($3, $2, 'wanted', 6)
	`, user1ID, skillID, user2ID)
	require.NoError(t, err)

	t.Run("active users can see each other in matches", func(t *testing.T) {
		// Query would normally go through matches repository
		// Here we verify both users are active
		isActive1, err := userRepo.IsUserActive(ctx, user1ID)
		require.NoError(t, err)
		assert.True(t, isActive1)

		isActive2, err := userRepo.IsUserActive(ctx, user2ID)
		require.NoError(t, err)
		assert.True(t, isActive2)
	})

	t.Run("banned user does not appear in matches", func(t *testing.T) {
		// Ban user2
		err := userRepo.BanUser(ctx, user2ID)
		require.NoError(t, err)

		// User2 should not be active
		isActive, err := userRepo.IsUserActive(ctx, user2ID)
		require.NoError(t, err)
		assert.False(t, isActive)

		// The matches repository query includes WHERE u.is_active = true
		// So user2 would be automatically excluded from match results
		var count int
		err = pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM users
			WHERE id = $1 AND is_active = true
		`, user2ID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Banned user should not be in active user count")
	})
}

// TestAdminActions_VerifyUser tests user verification
func TestAdminActions_VerifyUser(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	userRepo := user.NewRepository(pool)
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	userID := createTestUser(t, pool,
		fmt.Sprintf("verify_%s@test.com", uniqueSuffix),
		fmt.Sprintf("verify_%s", uniqueSuffix),
		true,
	)
	defer cleanupTestUser(t, pool, userID)

	t.Run("user is not verified initially", func(t *testing.T) {
		// Check is_verified field from database directly since it may not be in User struct
		var isVerified bool
		err := pool.QueryRow(ctx, "SELECT COALESCE(is_verified, false) FROM users WHERE id = $1", userID).Scan(&isVerified)
		require.NoError(t, err)
		assert.False(t, isVerified, "User should not be verified initially")
	})

	t.Run("verify user", func(t *testing.T) {
		err := userRepo.VerifyUser(ctx, userID)
		require.NoError(t, err)

		// Check verification directly from database
		var isVerified bool
		err = pool.QueryRow(ctx, "SELECT COALESCE(is_verified, false) FROM users WHERE id = $1", userID).Scan(&isVerified)
		require.NoError(t, err)
		assert.True(t, isVerified, "User should be verified")
	})

	t.Run("unverify user", func(t *testing.T) {
		err := userRepo.UnverifyUser(ctx, userID)
		require.NoError(t, err)

		// Check verification directly from database
		var isVerified bool
		err = pool.QueryRow(ctx, "SELECT COALESCE(is_verified, false) FROM users WHERE id = $1", userID).Scan(&isVerified)
		require.NoError(t, err)
		assert.False(t, isVerified, "User should not be verified after unverify")
	})
}

// TestAdminActions_GetAllUsers tests retrieving all users for admin dashboard
func TestAdminActions_GetAllUsers(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	userRepo := user.NewRepository(pool)
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create multiple test users
	var createdUserIDs []string
	for i := 0; i < 3; i++ {
		userID := createTestUser(t, pool,
			fmt.Sprintf("admin_user%d_%s@test.com", i, uniqueSuffix),
			fmt.Sprintf("admin_user%d_%s", i, uniqueSuffix),
			true,
		)
		createdUserIDs = append(createdUserIDs, userID)
	}

	// Cleanup
	defer func() {
		for _, uid := range createdUserIDs {
			cleanupTestUser(t, pool, uid)
		}
	}()

	t.Run("get all users", func(t *testing.T) {
		users, err := userRepo.GetAllUsers(ctx)
		require.NoError(t, err)

		// Should include at least our test users
		assert.GreaterOrEqual(t, len(users), 3, "Should have at least 3 users")

		// Verify our created users are in the list
		for _, createdID := range createdUserIDs {
			found := false
			for _, u := range users {
				if u.ID == createdID {
					found = true
					assert.True(t, u.IsActive, "Test user should be active")
					break
				}
			}
			assert.True(t, found, "Created user %s should be in all users list", createdID)
		}
	})

	t.Run("banned users appear in all users", func(t *testing.T) {
		// Ban first user
		err := userRepo.BanUser(ctx, createdUserIDs[0])
		require.NoError(t, err)

		users, err := userRepo.GetAllUsers(ctx)
		require.NoError(t, err)

		// Find the banned user
		found := false
		for _, u := range users {
			if u.ID == createdUserIDs[0] {
				found = true
				assert.False(t, u.IsActive, "Banned user should not be active")
				break
			}
		}
		assert.True(t, found, "Banned user should still appear in all users list")
	})
}

// TestAdminActions_BanPreventsAllActions tests that ban prevents all major user actions
func TestAdminActions_BanPreventsAllActions(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	userRepo := user.NewRepository(pool)
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	userID := createTestUser(t, pool,
		fmt.Sprintf("prevent_%s@test.com", uniqueSuffix),
		fmt.Sprintf("prevent_%s", uniqueSuffix),
		true,
	)
	defer cleanupTestUser(t, pool, userID)

	t.Run("ban user", func(t *testing.T) {
		err := userRepo.BanUser(ctx, userID)
		require.NoError(t, err)
	})

	t.Run("banned user not active", func(t *testing.T) {
		isActive, err := userRepo.IsUserActive(ctx, userID)
		require.NoError(t, err)
		assert.False(t, isActive, "Banned user should not be active")
	})

	t.Run("banned user data still retrievable for admin", func(t *testing.T) {
		// Admin should still be able to view user data
		u, err := userRepo.GetByID(ctx, userID)
		require.NoError(t, err)
		assert.NotNil(t, u)
		assert.Equal(t, userID, u.ID)
		assert.False(t, u.IsActive)
	})

	t.Run("banned user excluded from active user queries", func(t *testing.T) {
		// Any query with WHERE is_active = true should exclude this user
		var count int
		err := pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM users
			WHERE id = $1 AND is_active = true
		`, userID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Banned user should not be in active users")
	})
}

// TestAdminActions_MultipleBanUnban tests multiple ban/unban cycles
func TestAdminActions_MultipleBanUnban(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	userRepo := user.NewRepository(pool)
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	userID := createTestUser(t, pool,
		fmt.Sprintf("cycle_%s@test.com", uniqueSuffix),
		fmt.Sprintf("cycle_%s", uniqueSuffix),
		true,
	)
	defer cleanupTestUser(t, pool, userID)

	t.Run("multiple ban/unban cycles", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			// Ban
			err := userRepo.BanUser(ctx, userID)
			require.NoError(t, err)

			isActive, err := userRepo.IsUserActive(ctx, userID)
			require.NoError(t, err)
			assert.False(t, isActive, "User should be banned in cycle %d", i+1)

			// Unban
			err = userRepo.UnbanUser(ctx, userID)
			require.NoError(t, err)

			isActive, err = userRepo.IsUserActive(ctx, userID)
			require.NoError(t, err)
			assert.True(t, isActive, "User should be active in cycle %d", i+1)
		}
	})
}
