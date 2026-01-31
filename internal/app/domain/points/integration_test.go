package points

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestRepositoryIntegration tests the repository with a real database
// Run with: go test -v -tags=integration ./internal/app/domain/points/...
func TestRepositoryIntegration(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping integration test: DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool)

	// Create a test user for our tests
	testUserID := uuid.New()

	// First, we need to create a test user in the users table
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, email, username, hashed_password, display_name, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
		ON CONFLICT (id) DO NOTHING
	`, testUserID, "test_points_"+testUserID.String()[:8]+"@test.com", "testpoints"+testUserID.String()[:8], "hashedpw", "Test Points User")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Cleanup after test
	defer func() {
		if _, err := pool.Exec(ctx, "DELETE FROM points_history WHERE user_id = $1", testUserID); err != nil {
			t.Logf("failed to cleanup points_history: %v", err)
		}
		if _, err := pool.Exec(ctx, "DELETE FROM user_stats WHERE user_id = $1", testUserID); err != nil {
			t.Logf("failed to cleanup user_stats: %v", err)
		}
		if _, err := pool.Exec(ctx, "DELETE FROM users WHERE id = $1", testUserID); err != nil {
			t.Logf("failed to cleanup users: %v", err)
		}
	}()

	t.Run("EnsureUserStats", func(t *testing.T) {
		err := repo.EnsureUserStats(ctx, testUserID)
		if err != nil {
			t.Errorf("EnsureUserStats failed: %v", err)
		}
	})

	t.Run("GetUserPoints_NewUser", func(t *testing.T) {
		up, err := repo.GetUserPoints(ctx, testUserID)
		if err != nil {
			t.Errorf("GetUserPoints failed: %v", err)
		}
		if up.Points != 0 {
			t.Errorf("New user points = %d, want 0", up.Points)
		}
		if up.Tier != TierBronze {
			t.Errorf("New user tier = %s, want Bronze", up.Tier)
		}
	})

	t.Run("AddPoints_NoUpgrade", func(t *testing.T) {
		upgrade, err := repo.AddPoints(ctx, testUserID, 50, ActionSessionCompleted)
		if err != nil {
			t.Errorf("AddPoints failed: %v", err)
		}
		if upgrade != nil {
			t.Error("Expected no upgrade for 50 points")
		}

		up, _ := repo.GetUserPoints(ctx, testUserID)
		if up.Points != 50 {
			t.Errorf("Points = %d, want 50", up.Points)
		}
		if up.Tier != TierBronze {
			t.Errorf("Tier = %s, want Bronze", up.Tier)
		}
	})

	t.Run("AddPoints_UpgradeToSilver", func(t *testing.T) {
		// Add enough to cross 100 threshold
		upgrade, err := repo.AddPoints(ctx, testUserID, 60, ActionReviewReceived)
		if err != nil {
			t.Errorf("AddPoints failed: %v", err)
		}
		if upgrade == nil {
			t.Error("Expected upgrade to Silver")
		} else {
			if upgrade.OldTier != TierBronze {
				t.Errorf("OldTier = %s, want Bronze", upgrade.OldTier)
			}
			if upgrade.NewTier != TierSilver {
				t.Errorf("NewTier = %s, want Silver", upgrade.NewTier)
			}
			if upgrade.Points != 110 {
				t.Errorf("Points = %d, want 110", upgrade.Points)
			}
		}
	})

	t.Run("AddPoints_UpgradeToGold", func(t *testing.T) {
		// Add enough to cross 500 threshold
		upgrade, err := repo.AddPoints(ctx, testUserID, 400, ActionProfileComplete)
		if err != nil {
			t.Errorf("AddPoints failed: %v", err)
		}
		if upgrade == nil {
			t.Error("Expected upgrade to Gold")
		} else {
			if upgrade.OldTier != TierSilver {
				t.Errorf("OldTier = %s, want Silver", upgrade.OldTier)
			}
			if upgrade.NewTier != TierGold {
				t.Errorf("NewTier = %s, want Gold", upgrade.NewTier)
			}
		}
	})

	t.Run("GetUsersByTier", func(t *testing.T) {
		users, err := repo.GetUsersByTier(ctx, TierGold)
		if err != nil {
			t.Errorf("GetUsersByTier failed: %v", err)
		}

		found := false
		for _, u := range users {
			if u == testUserID {
				found = true
				break
			}
		}
		if !found {
			t.Error("Test user not found in Gold tier")
		}
	})

	t.Run("GetLeaderboard", func(t *testing.T) {
		leaders, err := repo.GetLeaderboard(ctx, 10)
		if err != nil {
			t.Errorf("GetLeaderboard failed: %v", err)
		}
		if len(leaders) == 0 {
			t.Error("Expected at least one leader")
		}
	})
}

// TestServiceIntegration tests the service with database
func TestServiceIntegration(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping integration test: DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool)
	hub := NewMockNotificationHub()
	service := NewService(repo, hub)

	// Create a test user
	testUserID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, email, username, hashed_password, display_name, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
	`, testUserID, "svc_test_"+testUserID.String()[:8]+"@test.com", "svctest"+testUserID.String()[:8], "hashedpw", "Service Test User")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	defer func() {
		if _, err := pool.Exec(ctx, "DELETE FROM points_history WHERE user_id = $1", testUserID); err != nil {
			t.Logf("failed to cleanup points_history: %v", err)
		}
		if _, err := pool.Exec(ctx, "DELETE FROM user_stats WHERE user_id = $1", testUserID); err != nil {
			t.Logf("failed to cleanup user_stats: %v", err)
		}
		if _, err := pool.Exec(ctx, "DELETE FROM users WHERE id = $1", testUserID); err != nil {
			t.Logf("failed to cleanup users: %v", err)
		}
	}()

	t.Run("AwardSessionPoints", func(t *testing.T) {
		err := service.EnsureUserStats(ctx, testUserID)
		if err != nil {
			t.Errorf("EnsureUserStats failed: %v", err)
		}

		upgrade, err := service.AwardSessionPoints(ctx, testUserID)
		if err != nil {
			t.Errorf("AwardSessionPoints failed: %v", err)
		}
		if upgrade != nil {
			t.Error("Unexpected upgrade from single session")
		}

		up, _ := service.GetUserPoints(ctx, testUserID)
		if up.Points != PointsReward[ActionSessionCompleted] {
			t.Errorf("Points = %d, want %d", up.Points, PointsReward[ActionSessionCompleted])
		}
	})

	t.Run("AwardReviewPoints_HighRating", func(t *testing.T) {
		hub.Clear()

		// Award multiple high-rating reviews to trigger upgrade
		for i := 0; i < 5; i++ {
			_, _ = service.AwardReviewPoints(ctx, testUserID, 5)
		}

		up, _ := service.GetUserPoints(ctx, testUserID)
		// Should have: 20 (session) + 5 * 20 (reviews with 2x multiplier) = 120
		expectedMin := 100 // At least silver threshold
		if up.Points < expectedMin {
			t.Errorf("Points = %d, want at least %d", up.Points, expectedMin)
		}

		// Should have received at least one upgrade notification
		if hub.GetMessageCount() < 1 {
			t.Error("Expected at least one upgrade notification")
		}
	})

	t.Run("GetProgressToNextTier", func(t *testing.T) {
		currentPoints, currentTier, pointsNeeded, nextTier, err := service.GetProgressToNextTier(ctx, testUserID)
		if err != nil {
			t.Errorf("GetProgressToNextTier failed: %v", err)
		}

		t.Logf("Progress: %d points, %s tier, %d to %s", currentPoints, currentTier, pointsNeeded, nextTier)

		if currentPoints <= 0 {
			t.Error("Expected positive points")
		}
	})
}

// BenchmarkCalculateTierFromPoints benchmarks tier calculation
func BenchmarkCalculateTierFromPoints(b *testing.B) {
	points := []int{0, 50, 100, 300, 500, 1000}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, p := range points {
			_ = CalculateTierFromPoints(p)
		}
	}
}

// BenchmarkPointsToNextTier benchmarks next tier calculation
func BenchmarkPointsToNextTier(b *testing.B) {
	testCases := []struct {
		points int
		tier   Tier
	}{
		{50, TierBronze},
		{200, TierSilver},
		{600, TierGold},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			_ = PointsToNextTier(tc.points, tc.tier)
		}
	}
}
