package review

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/badges"
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

// createTestSession creates a test session
func createTestSession(t *testing.T, pool *pgxpool.Pool, initiatorID, partnerID string, status string) string {
	ctx := context.Background()

	query := `
		INSERT INTO sessions (
			initiator_id, partner_id,
			initiator_offers, partner_offers,
			scheduled_start, scheduled_end,
			status
		) VALUES (
			$1, $2, ARRAY[]::TEXT[], ARRAY[]::TEXT[],
			NOW(), NOW() + interval '1 hour',
			$3
		) RETURNING id
	`

	var sessionID string
	err := pool.QueryRow(ctx, query, initiatorID, partnerID, status).Scan(&sessionID)
	require.NoError(t, err)

	return sessionID
}

// cleanupTestData removes test data
func cleanupTestData(t *testing.T, pool *pgxpool.Pool, userIDs ...string) {
	ctx := context.Background()
	for _, userID := range userIDs {
		_, _ = pool.Exec(ctx, "DELETE FROM reviews WHERE requester_id = $1 OR reviewer_id = $1", userID)
		_, _ = pool.Exec(ctx, "DELETE FROM sessions WHERE initiator_id = $1 OR partner_id = $1", userID)
		_, _ = pool.Exec(ctx, "DELETE FROM user_stats WHERE user_id = $1", userID)
		_, _ = pool.Exec(ctx, "DELETE FROM points_history WHERE user_id = $1", userID)
		_, _ = pool.Exec(ctx, "DELETE FROM user_badges WHERE user_id = $1", userID)
		_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	}
}

// TestReviewFlow_Integration tests the complete review submission workflow
func TestReviewFlow_Integration(t *testing.T) {
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
		fmt.Sprintf("reviewer_%s@test.com", uniqueSuffix),
		fmt.Sprintf("reviewer_%s", uniqueSuffix),
	)
	user2ID := createTestUser(t, pool,
		fmt.Sprintf("reviewee_%s@test.com", uniqueSuffix),
		fmt.Sprintf("reviewee_%s", uniqueSuffix),
	)
	defer cleanupTestData(t, pool, user1ID, user2ID)

	// Create a completed session
	sessionID := createTestSession(t, pool, user1ID, user2ID, "completed")

	t.Run("submit review successfully", func(t *testing.T) {
		err := repo.CreateSessionReview(ctx, user1ID, user2ID, sessionID, 5, "Great session!")
		require.NoError(t, err)

		// Verify review exists
		exists, err := repo.ReviewExistsForSession(ctx, user1ID, sessionID)
		require.NoError(t, err)
		assert.True(t, exists, "Review should exist after submission")
	})

	t.Run("verify review stats updated", func(t *testing.T) {
		totalReviews, avgRating, err := repo.GetStats(ctx, uuid.MustParse(user2ID))
		require.NoError(t, err)

		assert.Equal(t, 1, totalReviews, "Should have 1 review")
		assert.Equal(t, 5.0, avgRating, "Average rating should be 5.0")
	})

	t.Run("get reviews received", func(t *testing.T) {
		reviews, err := repo.GetReviewsReceived(ctx, uuid.MustParse(user2ID))
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(reviews), 1, "Should have at least one review")
		if len(reviews) > 0 {
			assert.Equal(t, 5, reviews[0].Rating, "Rating should be 5")
			assert.Equal(t, "Great session!", reviews[0].Comment)
		}
	})

	t.Run("prevent duplicate review submission", func(t *testing.T) {
		// Try to submit another review for the same session
		_ = repo.CreateSessionReview(ctx, user1ID, user2ID, sessionID, 4, "Another review")
		// Should succeed (creates new record), but validation should prevent this in handler

		// Check validation
		err := repo.ValidateReviewSubmission(ctx, user1ID, user2ID, sessionID)
		assert.ErrorIs(t, err, ErrDuplicateReview, "Should return duplicate review error")
	})

	t.Run("prevent review on pending session", func(t *testing.T) {
		pendingSessionID := createTestSession(t, pool, user1ID, user2ID, "pending")

		err := repo.ValidateReviewSubmission(ctx, user1ID, user2ID, pendingSessionID)
		assert.ErrorIs(t, err, ErrSessionNotCompleted, "Should not allow review on pending session")
	})

	t.Run("prevent self-review", func(t *testing.T) {
		err := repo.ValidateReviewSubmission(ctx, user1ID, user1ID, sessionID)
		assert.ErrorIs(t, err, ErrCannotReviewSelf, "Should not allow self-review")
	})

	t.Run("prevent review from non-participant", func(t *testing.T) {
		user3ID := createTestUser(t, pool,
			fmt.Sprintf("nonparticipant_%s@test.com", uniqueSuffix),
			fmt.Sprintf("nonparticipant_%s", uniqueSuffix),
		)
		defer cleanupTestData(t, pool, user3ID)

		err := repo.ValidateReviewSubmission(ctx, user3ID, user2ID, sessionID)
		assert.ErrorIs(t, err, ErrNotSessionParticipant, "Should not allow non-participant to review")
	})

	t.Run("prevent review on non-existent session", func(t *testing.T) {
		fakeSessionID := uuid.New().String()
		err := repo.ValidateReviewSubmission(ctx, user1ID, user2ID, fakeSessionID)
		assert.ErrorIs(t, err, ErrSessionNotFound, "Should return session not found error")
	})
}

// TestReviewFlow_PointsAwarded tests points are awarded for reviews
func TestReviewFlow_PointsAwarded(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	reviewRepo := NewRepository(pool)
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

	// Ensure user stats exist
	err := pointsService.EnsureUserStats(ctx, uuid.MustParse(user2ID))
	require.NoError(t, err)

	// Create a completed session
	sessionID := createTestSession(t, pool, user1ID, user2ID, "completed")

	t.Run("5-star review awards double points", func(t *testing.T) {
		// Reset user2's points to 0 for this test
		_, _ = pool.Exec(ctx, "UPDATE user_stats SET points = 0, tier = 'Bronze' WHERE user_id = $1", user2ID)

		// Get initial points (should be 0)
		initialPoints, err := pointsService.GetUserPoints(ctx, uuid.MustParse(user2ID))
		require.NoError(t, err)
		initialPointsValue := initialPoints.Points

		// Submit 5-star review
		err = reviewRepo.CreateSessionReview(ctx, user1ID, user2ID, sessionID, 5, "Excellent!")
		require.NoError(t, err)

		// Award points
		upgrade, err := pointsService.AwardReviewPoints(ctx, uuid.MustParse(user2ID), 5)
		require.NoError(t, err)

		// Get updated points
		updatedPoints, err := pointsService.GetUserPoints(ctx, uuid.MustParse(user2ID))
		require.NoError(t, err)

		// Should have awarded double points for 5-star review
		pointsDelta := updatedPoints.Points - initialPointsValue
		expectedDelta := points.PointsReward[points.ActionReviewReceived] * 2
		assert.Equal(t, expectedDelta, pointsDelta, "Should award double points for 5-star review")

		// Tier upgrade might have occurred
		if upgrade != nil {
			t.Logf("Tier upgraded from %s to %s", upgrade.OldTier, upgrade.NewTier)
		}
	})

	t.Run("4-star review awards standard points", func(t *testing.T) {
		// Create another session
		sessionID2 := createTestSession(t, pool, user2ID, user1ID, "completed")

		// Ensure stats exist and reset points to 0 for this test
		err = pointsService.EnsureUserStats(ctx, uuid.MustParse(user1ID))
		require.NoError(t, err)
		_, _ = pool.Exec(ctx, "UPDATE user_stats SET points = 0, tier = 'Bronze' WHERE user_id = $1", user1ID)

		// Get initial points (should be 0)
		initialPoints, err := pointsService.GetUserPoints(ctx, uuid.MustParse(user1ID))
		require.NoError(t, err)
		initialPointsValue := initialPoints.Points

		// Submit 4-star review
		err = reviewRepo.CreateSessionReview(ctx, user2ID, user1ID, sessionID2, 4, "Good session")
		require.NoError(t, err)

		// Award points
		_, err = pointsService.AwardReviewPoints(ctx, uuid.MustParse(user1ID), 4)
		require.NoError(t, err)

		// Get updated points
		updatedPoints, err := pointsService.GetUserPoints(ctx, uuid.MustParse(user1ID))
		require.NoError(t, err)

		// Should have awarded standard points for 4-star review
		pointsDelta := updatedPoints.Points - initialPointsValue
		expectedDelta := points.PointsReward[points.ActionReviewReceived]
		assert.Equal(t, expectedDelta, pointsDelta, "Should award standard points for 4-star review")
	})

	t.Run("low rating awards no points", func(t *testing.T) {
		// Create another user
		user3ID := createTestUser(t, pool,
			fmt.Sprintf("points3_%s@test.com", uniqueSuffix),
			fmt.Sprintf("points3_%s", uniqueSuffix),
		)
		defer cleanupTestData(t, pool, user3ID)

		// Ensure stats exist
		err := pointsService.EnsureUserStats(ctx, uuid.MustParse(user3ID))
		require.NoError(t, err)

		// Get initial points
		initialPoints, err := pointsService.GetUserPoints(ctx, uuid.MustParse(user3ID))
		require.NoError(t, err)

		// Award points for 3-star review (should not award)
		_, err = pointsService.AwardReviewPoints(ctx, uuid.MustParse(user3ID), 3)
		require.NoError(t, err)

		// Get updated points
		updatedPoints, err := pointsService.GetUserPoints(ctx, uuid.MustParse(user3ID))
		require.NoError(t, err)

		// Points should remain the same
		assert.Equal(t, initialPoints.Points, updatedPoints.Points, "Should not award points for rating below 4")
	})
}

// TestReviewFlow_TopTeacherBadge tests top teacher badge awarding
func TestReviewFlow_TopTeacherBadge(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	reviewRepo := NewRepository(pool)
	badgeRepo := badges.NewRepository(pool)
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create teacher and multiple students
	teacherID := createTestUser(t, pool,
		fmt.Sprintf("teacher_%s@test.com", uniqueSuffix),
		fmt.Sprintf("teacher_%s", uniqueSuffix),
	)
	defer cleanupTestData(t, pool, teacherID)

	t.Run("award top teacher badge after 5 high reviews", func(t *testing.T) {
		// Create 5 students and sessions
		for i := 0; i < 5; i++ {
			studentID := createTestUser(t, pool,
				fmt.Sprintf("student%d_%s@test.com", i, uniqueSuffix),
				fmt.Sprintf("student%d_%s", i, uniqueSuffix),
			)
			defer cleanupTestData(t, pool, studentID)

			sessionID := createTestSession(t, pool, studentID, teacherID, "completed")

			// Submit 5-star review
			err := reviewRepo.CreateSessionReview(ctx, studentID, teacherID, sessionID, 5, "Amazing teacher!")
			require.NoError(t, err)
		}

		// Check stats
		totalReviews, avgRating, err := reviewRepo.GetStats(ctx, uuid.MustParse(teacherID))
		require.NoError(t, err)

		assert.Equal(t, 5, totalReviews, "Should have 5 reviews")
		assert.Equal(t, 5.0, avgRating, "Average should be 5.0")

		// Award badge (simulating handler logic)
		if totalReviews >= 5 && avgRating >= 4.5 {
			err = badgeRepo.AwardBadge(ctx, uuid.MustParse(teacherID), "top_teacher")
			require.NoError(t, err)
		}

		// Verify badge was awarded
		hasBadge, err := badgeRepo.HasBadge(ctx, uuid.MustParse(teacherID), "top_teacher")
		require.NoError(t, err)
		assert.True(t, hasBadge, "Should have top_teacher badge")
	})
}

// TestReviewFlow_MultipleSessionReviews tests reviewing multiple sessions
func TestReviewFlow_MultipleSessionReviews(t *testing.T) {
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

	t.Run("can review multiple sessions with same user", func(t *testing.T) {
		// Create 3 completed sessions
		for i := 0; i < 3; i++ {
			sessionID := createTestSession(t, pool, user1ID, user2ID, "completed")

			// Submit review
			err := repo.CreateSessionReview(ctx, user1ID, user2ID, sessionID, 5, fmt.Sprintf("Session %d", i+1))
			require.NoError(t, err)

			// Verify review exists
			exists, err := repo.ReviewExistsForSession(ctx, user1ID, sessionID)
			require.NoError(t, err)
			assert.True(t, exists, "Review should exist for session %d", i+1)
		}

		// Check total reviews
		totalReviews, avgRating, err := repo.GetStats(ctx, uuid.MustParse(user2ID))
		require.NoError(t, err)

		assert.Equal(t, 3, totalReviews, "Should have 3 reviews")
		assert.Equal(t, 5.0, avgRating, "Average should be 5.0")
	})

	t.Run("verify reviews received list", func(t *testing.T) {
		reviews, err := repo.GetReviewsReceived(ctx, uuid.MustParse(user2ID))
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(reviews), 3, "Should have at least 3 reviews")
	})
}

// TestReviewFlow_ValidationEdgeCases tests various validation edge cases
func TestReviewFlow_ValidationEdgeCases(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewRepository(pool)
	ctx := context.Background()
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	user1ID := createTestUser(t, pool,
		fmt.Sprintf("edge1_%s@test.com", uniqueSuffix),
		fmt.Sprintf("edge1_%s", uniqueSuffix),
	)
	user2ID := createTestUser(t, pool,
		fmt.Sprintf("edge2_%s@test.com", uniqueSuffix),
		fmt.Sprintf("edge2_%s", uniqueSuffix),
	)
	user3ID := createTestUser(t, pool,
		fmt.Sprintf("edge3_%s@test.com", uniqueSuffix),
		fmt.Sprintf("edge3_%s", uniqueSuffix),
	)
	defer cleanupTestData(t, pool, user1ID, user2ID, user3ID)

	sessionID := createTestSession(t, pool, user1ID, user2ID, "completed")

	t.Run("wrong target user validation", func(t *testing.T) {
		// User1 tries to review User3 for a session with User2
		err := repo.ValidateReviewSubmission(ctx, user1ID, user3ID, sessionID)
		assert.ErrorIs(t, err, ErrNotSessionParticipant, "Should reject review for non-participant")
	})

	t.Run("cancelled session cannot be reviewed", func(t *testing.T) {
		cancelledSessionID := createTestSession(t, pool, user1ID, user2ID, "cancelled")

		err := repo.ValidateReviewSubmission(ctx, user1ID, user2ID, cancelledSessionID)
		assert.ErrorIs(t, err, ErrSessionNotCompleted, "Should not allow review on cancelled session")
	})

	t.Run("both participants can review each other", func(t *testing.T) {
		session2ID := createTestSession(t, pool, user1ID, user2ID, "completed")

		// User1 reviews User2
		err := repo.CreateSessionReview(ctx, user1ID, user2ID, session2ID, 5, "Great!")
		require.NoError(t, err)

		// User2 reviews User1 (should also work)
		err = repo.CreateSessionReview(ctx, user2ID, user1ID, session2ID, 4, "Good!")
		require.NoError(t, err)

		// Both reviews should exist
		exists1, err := repo.ReviewExistsForSession(ctx, user1ID, session2ID)
		require.NoError(t, err)
		assert.True(t, exists1)

		exists2, err := repo.ReviewExistsForSession(ctx, user2ID, session2ID)
		require.NoError(t, err)
		assert.True(t, exists2)
	})
}
