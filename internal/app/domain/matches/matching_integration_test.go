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

// getIntegrationTestDBPool creates a connection pool to the test database
func getIntegrationTestDBPool(t *testing.T) *pgxpool.Pool {
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

// createIntegrationTestUser creates a test user with skills
func createIntegrationTestUser(t *testing.T, pool *pgxpool.Pool, email, username string) string {
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

// addUserSkill adds a skill to a user
func addUserSkill(t *testing.T, pool *pgxpool.Pool, userID, skillName, skillType string, proficiency int) {
	ctx := context.Background()

	// Get or create a category
	var categoryID uuid.UUID
	err := pool.QueryRow(ctx, `
		SELECT id FROM skill_categories LIMIT 1
	`).Scan(&categoryID)
	if err != nil {
		// Create a test category if none exists
		err = pool.QueryRow(ctx, `
			INSERT INTO skill_categories (name, description)
			VALUES ('Technology', 'Technology skills')
			ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`).Scan(&categoryID)
		require.NoError(t, err)
	}

	// Get or create skill
	var skillID uuid.UUID
	err = pool.QueryRow(ctx, `
		INSERT INTO skills (name, category_id)
		VALUES ($1, $2)
		ON CONFLICT (category_id, name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, skillName, categoryID).Scan(&skillID)
	require.NoError(t, err)

	// Add user skill
	_, err = pool.Exec(ctx, `
		INSERT INTO user_skills (user_id, skill_id, skill_type, proficiency)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, skill_id, skill_type) DO UPDATE SET proficiency = EXCLUDED.proficiency
	`, userID, skillID, skillType, proficiency)
	require.NoError(t, err)
}

// cleanupIntegrationTestUser removes test user and associated data
func cleanupIntegrationTestUser(t *testing.T, pool *pgxpool.Pool, userID string) {
	ctx := context.Background()
	_, _ = pool.Exec(ctx, "DELETE FROM user_skills WHERE user_id = $1", userID)
	_, _ = pool.Exec(ctx, "DELETE FROM match_history WHERE user_id_a = $1 OR user_id_b = $1", userID)
	_, _ = pool.Exec(ctx, "DELETE FROM user_badges WHERE user_id = $1", userID)
	_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
}

// TestMatchingFlow_Integration tests the complete matching workflow
func TestMatchingFlow_Integration(t *testing.T) {
	pool := getIntegrationTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewRepository(pool)
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create two users with complementary skills
	user1ID := createIntegrationTestUser(t, pool,
		fmt.Sprintf("user1_%s@test.com", uniqueSuffix),
		fmt.Sprintf("user1_%s", uniqueSuffix),
	)
	defer cleanupIntegrationTestUser(t, pool, user1ID)

	user2ID := createIntegrationTestUser(t, pool,
		fmt.Sprintf("user2_%s@test.com", uniqueSuffix),
		fmt.Sprintf("user2_%s", uniqueSuffix),
	)
	defer cleanupIntegrationTestUser(t, pool, user2ID)

	// User1 offers Go, wants Python
	addUserSkill(t, pool, user1ID, "Go", "offered", 8)
	addUserSkill(t, pool, user1ID, "Python", "wanted", 5)

	// User2 offers Python, wants Go (perfect match)
	addUserSkill(t, pool, user2ID, "Python", "offered", 7)
	addUserSkill(t, pool, user2ID, "Go", "wanted", 6)

	t.Run("generate matches", func(t *testing.T) {
		matches, err := repo.GetPendingMatches(ctx, user1ID, 10)
		require.NoError(t, err)

		// Should find user2 as a match
		found := false
		var matchScore float64
		for _, m := range matches {
			if m.UserID == user2ID {
				found = true
				matchScore = m.MatchScore
				assert.Greater(t, m.MatchScore, 0.0, "Match score should be positive")
				assert.NotEmpty(t, m.OverlapSkills, "Should have overlap skills")
				break
			}
		}
		assert.True(t, found, "User2 should appear in User1's matches")
		t.Logf("Match score: %.2f", matchScore)
	})

	t.Run("verify overlap skills", func(t *testing.T) {
		overlapSkills, err := repo.GetOverlapSkills(ctx, user1ID, user2ID)
		require.NoError(t, err)

		// Should have 2 overlap skills (Go and Python)
		assert.GreaterOrEqual(t, len(overlapSkills), 1, "Should have at least one overlap skill")

		// Verify skill directions
		for _, skill := range overlapSkills {
			assert.NotEmpty(t, skill.Name)
			t.Logf("Overlap skill: %s (they teach: %v)", skill.Name, skill.TheyTeach)
		}
	})

	t.Run("record match action - accept", func(t *testing.T) {
		err := repo.RecordMatchAction(ctx, user1ID, user2ID, true)
		require.NoError(t, err)

		// Verify match history recorded
		var count int
		err = pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM match_history
			WHERE user_id_a = $1 AND user_id_b = $2 AND interaction_initiated = true
		`, user1ID, user2ID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Match action should be recorded")
	})

	t.Run("check not yet connected", func(t *testing.T) {
		connected, err := repo.AreConnected(ctx, user1ID, user2ID)
		require.NoError(t, err)
		assert.False(t, connected, "Should not be connected yet (only one-way)")
	})

	t.Run("mutual match creates connection", func(t *testing.T) {
		// User2 also accepts User1
		err := repo.RecordMatchAction(ctx, user2ID, user1ID, true)
		require.NoError(t, err)

		// Now they should be connected
		connected, err := repo.AreConnected(ctx, user1ID, user2ID)
		require.NoError(t, err)
		assert.True(t, connected, "Should be connected after mutual acceptance")

		// Check mutual matches list
		mutualMatches, err := repo.GetMutualMatches(ctx, user1ID, 10)
		require.NoError(t, err)

		found := false
		for _, m := range mutualMatches {
			if m.UserID == user2ID {
				found = true
				break
			}
		}
		assert.True(t, found, "User2 should appear in User1's mutual matches")
	})

	t.Run("verify both users see the connection", func(t *testing.T) {
		// User1's perspective
		mutualMatches1, err := repo.GetMutualMatches(ctx, user1ID, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(mutualMatches1), 1)

		// User2's perspective
		mutualMatches2, err := repo.GetMutualMatches(ctx, user2ID, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(mutualMatches2), 1)
	})

	t.Run("record match action - reject", func(t *testing.T) {
		// Create another user for rejection test
		user3ID := createIntegrationTestUser(t, pool,
			fmt.Sprintf("user3_%s@test.com", uniqueSuffix),
			fmt.Sprintf("user3_%s", uniqueSuffix),
		)
		defer cleanupIntegrationTestUser(t, pool, user3ID)

		// User1 rejects User3
		err := repo.RecordMatchAction(ctx, user1ID, user3ID, false)
		require.NoError(t, err)

		// User3 should not appear in future matches for User1
		matches, err := repo.GetPendingMatches(ctx, user1ID, 10)
		require.NoError(t, err)

		for _, m := range matches {
			assert.NotEqual(t, user3ID, m.UserID, "Rejected user should not appear in matches")
		}
	})

	t.Run("check connection count", func(t *testing.T) {
		count, err := repo.GetConnectionCount(ctx, user1ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1, "Should have at least one connection")
	})

	t.Run("has connections check", func(t *testing.T) {
		hasConns, err := repo.HasConnections(ctx, user1ID)
		require.NoError(t, err)
		assert.True(t, hasConns, "User should have connections")
	})
}

// TestMatchingFlow_FirstMatchBadge tests badge awarding on first match
func TestMatchingFlow_FirstMatchBadge(t *testing.T) {
	pool := getIntegrationTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	matchRepo := NewRepository(pool)
	badgeRepo := badges.NewRepository(pool)
	badgeService := badges.NewService(badgeRepo, nil) // No hub for test
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create two users
	user1ID := createIntegrationTestUser(t, pool,
		fmt.Sprintf("badge1_%s@test.com", uniqueSuffix),
		fmt.Sprintf("badge1_%s", uniqueSuffix),
	)
	defer cleanupIntegrationTestUser(t, pool, user1ID)

	user2ID := createIntegrationTestUser(t, pool,
		fmt.Sprintf("badge2_%s@test.com", uniqueSuffix),
		fmt.Sprintf("badge2_%s", uniqueSuffix),
	)
	defer cleanupIntegrationTestUser(t, pool, user2ID)

	// Add complementary skills
	addUserSkill(t, pool, user1ID, "JavaScript", "offered", 7)
	addUserSkill(t, pool, user2ID, "JavaScript", "wanted", 6)

	t.Run("no badge before match", func(t *testing.T) {
		hasBadge, err := badgeRepo.HasBadge(ctx, uuid.MustParse(user1ID), "first_match")
		require.NoError(t, err)
		assert.False(t, hasBadge, "Should not have first_match badge yet")
	})

	t.Run("create mutual match", func(t *testing.T) {
		// User1 accepts User2
		err := matchRepo.RecordMatchAction(ctx, user1ID, user2ID, true)
		require.NoError(t, err)

		// User2 accepts User1
		err = matchRepo.RecordMatchAction(ctx, user2ID, user1ID, true)
		require.NoError(t, err)

		// Verify connection
		connected, err := matchRepo.AreConnected(ctx, user1ID, user2ID)
		require.NoError(t, err)
		assert.True(t, connected)
	})

	t.Run("award first match badge", func(t *testing.T) {
		// Award badge to user1
		err := badgeService.AwardBadge(ctx, uuid.MustParse(user1ID), "first_match")
		require.NoError(t, err)

		// Verify badge was awarded
		hasBadge, err := badgeRepo.HasBadge(ctx, uuid.MustParse(user1ID), "first_match")
		require.NoError(t, err)
		assert.True(t, hasBadge, "Should have first_match badge after awarding")

		// Verify badge shows in user badges
		userBadges, err := badgeService.GetUserBadges(ctx, uuid.MustParse(user1ID))
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(userBadges), 1, "Should have at least one badge")
	})

	t.Run("duplicate badge award is idempotent", func(t *testing.T) {
		// Try to award again - should not error
		err := badgeService.AwardBadge(ctx, uuid.MustParse(user1ID), "first_match")
		require.NoError(t, err)

		// Still should have only one badge record
		userBadges, err := badgeService.GetUserBadges(ctx, uuid.MustParse(user1ID))
		require.NoError(t, err)

		firstMatchCount := 0
		for _, badge := range userBadges {
			if badge.BadgeCode == "first_match" {
				firstMatchCount++
			}
		}
		assert.Equal(t, 1, firstMatchCount, "Should have exactly one first_match badge")
	})
}

// TestMatchingFlow_PendingMatches tests pending sent and received matches
func TestMatchingFlow_PendingMatches(t *testing.T) {
	pool := getIntegrationTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewRepository(pool)
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create three users
	user1ID := createIntegrationTestUser(t, pool,
		fmt.Sprintf("pending1_%s@test.com", uniqueSuffix),
		fmt.Sprintf("pending1_%s", uniqueSuffix),
	)
	defer cleanupIntegrationTestUser(t, pool, user1ID)

	user2ID := createIntegrationTestUser(t, pool,
		fmt.Sprintf("pending2_%s@test.com", uniqueSuffix),
		fmt.Sprintf("pending2_%s", uniqueSuffix),
	)
	defer cleanupIntegrationTestUser(t, pool, user2ID)

	user3ID := createIntegrationTestUser(t, pool,
		fmt.Sprintf("pending3_%s@test.com", uniqueSuffix),
		fmt.Sprintf("pending3_%s", uniqueSuffix),
	)
	defer cleanupIntegrationTestUser(t, pool, user3ID)

	t.Run("user1 likes user2", func(t *testing.T) {
		err := repo.RecordMatchAction(ctx, user1ID, user2ID, true)
		require.NoError(t, err)

		// User1 should see user2 in pending sent
		pendingSent, err := repo.GetPendingSentMatches(ctx, user1ID, 10)
		require.NoError(t, err)

		found := false
		for _, m := range pendingSent {
			if m.UserID == user2ID {
				found = true
				break
			}
		}
		assert.True(t, found, "User2 should appear in User1's pending sent matches")
	})

	t.Run("user3 likes user1", func(t *testing.T) {
		err := repo.RecordMatchAction(ctx, user3ID, user1ID, true)
		require.NoError(t, err)

		// User1 should see user3 in pending received
		pendingReceived, err := repo.GetPendingReceivedMatches(ctx, user1ID, 10)
		require.NoError(t, err)

		found := false
		for _, m := range pendingReceived {
			if m.UserID == user3ID {
				found = true
				break
			}
		}
		assert.True(t, found, "User3 should appear in User1's pending received matches")
	})

	t.Run("after user1 accepts user3, no longer pending", func(t *testing.T) {
		err := repo.RecordMatchAction(ctx, user1ID, user3ID, true)
		require.NoError(t, err)

		// Should now be connected
		connected, err := repo.AreConnected(ctx, user1ID, user3ID)
		require.NoError(t, err)
		assert.True(t, connected)

		// User3 should not appear in pending received anymore
		pendingReceived, err := repo.GetPendingReceivedMatches(ctx, user1ID, 10)
		require.NoError(t, err)

		for _, m := range pendingReceived {
			assert.NotEqual(t, user3ID, m.UserID, "User3 should not be in pending received after acceptance")
		}
	})
}

// TestMatchingFlow_NoSelfMatch tests that users don't match with themselves
func TestMatchingFlow_NoSelfMatch(t *testing.T) {
	pool := getIntegrationTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewRepository(pool)
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	userID := createIntegrationTestUser(t, pool,
		fmt.Sprintf("self_%s@test.com", uniqueSuffix),
		fmt.Sprintf("self_%s", uniqueSuffix),
	)
	defer cleanupIntegrationTestUser(t, pool, userID)

	addUserSkill(t, pool, userID, "React", "offered", 8)
	addUserSkill(t, pool, userID, "React", "wanted", 5)

	matches, err := repo.GetPendingMatches(ctx, userID, 10)
	require.NoError(t, err)

	// User should not match with themselves
	for _, m := range matches {
		assert.NotEqual(t, userID, m.UserID, "User should not match with themselves")
	}
}
