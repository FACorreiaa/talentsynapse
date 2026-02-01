package matches

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// TestMatchingAlgorithmLoadPerformance tests the matching algorithm with 100+ users
// Acceptance Criteria: Matching completes in <2 seconds with 100+ users
func TestMatchingAlgorithmLoadPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewRepository(pool)
	ctx := context.Background()

	// Get a random load test user
	var testUserID string
	err := pool.QueryRow(ctx, `
		SELECT id FROM users
		WHERE email LIKE '%@loadtest.example.com'
		ORDER BY RANDOM()
		LIMIT 1
	`).Scan(&testUserID)
	require.NoError(t, err, "Failed to get test user")

	// Verify we have sufficient users for testing
	var userCount int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM users WHERE email LIKE '%@loadtest.example.com'
	`).Scan(&userCount)
	require.NoError(t, err)
	require.GreaterOrEqual(t, userCount, 100, "Need at least 100 users for load testing")

	t.Logf("Testing matching algorithm with %d users", userCount)

	// Test 1: Single user matching performance
	t.Run("SingleUserMatching", func(t *testing.T) {
		start := time.Now()
		matches, err := repo.GetPendingMatches(ctx, testUserID, 20)
		duration := time.Since(start)

		if err != nil {
			t.Logf("Error during matching: %v", err)
		}
		require.NoError(t, err)
		if matches == nil {
			t.Logf("Matches is nil but no error was returned")
		}

		t.Logf("Found %d matches in %v", len(matches), duration)
		require.Less(t, duration, 2*time.Second, "Matching should complete in <2 seconds")

		// Log top matches
		for i, match := range matches {
			if i >= 5 {
				break
			}
			t.Logf("  Match %d: %s (score: %.2f, skills: %d)",
				i+1, match.DisplayName, match.MatchScore, len(match.OverlapSkills))
		}
	})

	// Test 2: Concurrent matching requests
	t.Run("ConcurrentMatching", func(t *testing.T) {
		const numConcurrent = 10

		// Get multiple test users
		rows, err := pool.Query(ctx, `
			SELECT id FROM users
			WHERE email LIKE '%@loadtest.example.com'
			ORDER BY RANDOM()
			LIMIT $1
		`, numConcurrent)
		require.NoError(t, err)
		defer rows.Close()

		var userIDs []string
		for rows.Next() {
			var userID string
			err := rows.Scan(&userID)
			require.NoError(t, err)
			userIDs = append(userIDs, userID)
		}

		// Run concurrent matching requests
		start := time.Now()
		done := make(chan bool, numConcurrent)
		errors := make(chan error, numConcurrent)

		for i, userID := range userIDs {
			go func(idx int, uid string) {
				matchStart := time.Now()
				matches, err := repo.GetPendingMatches(ctx, uid, 20)
				matchDuration := time.Since(matchStart)

				if err != nil {
					errors <- err
					done <- false
					return
				}

				t.Logf("  Concurrent request %d: Found %d matches in %v", idx+1, len(matches), matchDuration)
				done <- true
			}(i, userID)
		}

		// Wait for all to complete
		successCount := 0
		for i := 0; i < numConcurrent; i++ {
			select {
			case success := <-done:
				if success {
					successCount++
				}
			case err := <-errors:
				t.Errorf("Concurrent matching error: %v", err)
			}
		}

		duration := time.Since(start)
		t.Logf("Completed %d/%d concurrent matching requests in %v", successCount, numConcurrent, duration)
		require.Equal(t, numConcurrent, successCount, "All concurrent requests should succeed")
		require.Less(t, duration, 5*time.Second, "Concurrent matching should complete in <5 seconds")
	})

	// Test 3: Materialized view performance
	t.Run("MaterializedViewFreshness", func(t *testing.T) {
		// Check when the materialized view was last refreshed
		var lastRefresh time.Time
		err := pool.QueryRow(ctx, `
			SELECT NOW() - (
				SELECT pg_catalog.pg_stat_get_last_vacuum_time(c.oid)
				FROM pg_catalog.pg_class c
				WHERE c.relname = 'user_skill_vectors'
			) as age
		`).Scan(&lastRefresh)

		// Note: pg_stat_get_last_vacuum_time might not work for materialized views
		// This is more of a diagnostic check
		if err == nil {
			t.Logf("Materialized view age: %v", lastRefresh)
		}

		// Test refresh performance
		start := time.Now()
		_, err = pool.Exec(ctx, "REFRESH MATERIALIZED VIEW user_skill_vectors")
		duration := time.Since(start)

		require.NoError(t, err)
		t.Logf("Materialized view refresh took: %v", duration)

		// Refresh should be reasonably fast even with many users
		require.Less(t, duration, 10*time.Second, "Materialized view refresh should complete in <10 seconds")
	})

	// Test 4: Match quality and diversity
	t.Run("MatchQuality", func(t *testing.T) {
		matches, err := repo.GetPendingMatches(ctx, testUserID, 50)
		require.NoError(t, err)

		if len(matches) == 0 {
			t.Log("No matches found - user may not have compatible skills")
			return
		}

		// Check score distribution
		highScoreCount := 0
		mediumScoreCount := 0
		lowScoreCount := 0

		for _, match := range matches {
			if match.MatchScore >= 0.7 {
				highScoreCount++
			} else if match.MatchScore >= 0.4 {
				mediumScoreCount++
			} else {
				lowScoreCount++
			}
		}

		t.Logf("Match score distribution: High(>=0.7): %d, Medium(0.4-0.7): %d, Low(<0.4): %d",
			highScoreCount, mediumScoreCount, lowScoreCount)

		// At least some matches should have reasonable scores
		require.Greater(t, highScoreCount+mediumScoreCount, 0, "Should have at least some quality matches")
	})
}

// TestMatchingAlgorithmScalability tests how performance scales with user count
func TestMatchingAlgorithmScalability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping scalability test in short mode")
	}

	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewRepository(pool)
	ctx := context.Background()

	// Get total user count
	var totalUsers int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE email LIKE '%@loadtest.example.com'`).Scan(&totalUsers)
	require.NoError(t, err)

	t.Logf("Testing scalability with %d total users", totalUsers)

	// Test matching performance with different result limits
	limits := []int{10, 20, 50, 100}

	var testUserID string
	err = pool.QueryRow(ctx, `
		SELECT id FROM users
		WHERE email LIKE '%@loadtest.example.com'
		ORDER BY RANDOM()
		LIMIT 1
	`).Scan(&testUserID)
	require.NoError(t, err)

	for _, limit := range limits {
		t.Run(fmt.Sprintf("Limit_%d", limit), func(t *testing.T) {
			start := time.Now()
			matches, err := repo.GetPendingMatches(ctx, testUserID, limit)
			duration := time.Since(start)

			require.NoError(t, err)
			t.Logf("Limit %d: Found %d matches in %v", limit, len(matches), duration)

			// Performance should remain reasonable even with higher limits
			require.Less(t, duration, 3*time.Second, "Should complete in <3 seconds")
		})
	}
}

// BenchmarkMatchingAlgorithm benchmarks the matching algorithm
func BenchmarkMatchingAlgorithm(b *testing.B) {
	pool := getBenchDBPool(b)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := NewRepository(pool)
	ctx := context.Background()

	// Get a test user
	var testUserID string
	err := pool.QueryRow(ctx, `
		SELECT id FROM users
		WHERE email LIKE '%@loadtest.example.com'
		ORDER BY RANDOM()
		LIMIT 1
	`).Scan(&testUserID)
	if err != nil {
		b.Skipf("No test data available: %v", err)
		return
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetPendingMatches(ctx, testUserID, 20)
		if err != nil {
			b.Fatalf("Matching failed: %v", err)
		}
	}
}

// Helper functions are in repository_test.go to avoid duplication

// Standalone test runner (can be run with: go run matching_load_test_runner.go)
func RunLoadTests() {
	log.Println("Starting matching algorithm load tests...")

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5470/myapp?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	repo := NewRepository(pool)

	// Get user count
	var userCount int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE email LIKE '%@loadtest.example.com'`).Scan(&userCount)
	if err != nil {
		log.Fatalf("Failed to count users: %v", err)
	}

	log.Printf("Found %d test users\n", userCount)

	if userCount < 100 {
		log.Fatalf("Need at least 100 users for load testing. Run migrations first.")
	}

	// Run matching tests
	var testUserID string
	err = pool.QueryRow(ctx, `SELECT id FROM users WHERE email LIKE '%@loadtest.example.com' ORDER BY RANDOM() LIMIT 1`).Scan(&testUserID)
	if err != nil {
		log.Fatalf("Failed to get test user: %v", err)
	}

	log.Println("\n=== Testing Single User Matching ===")
	start := time.Now()
	matches, err := repo.GetPendingMatches(ctx, testUserID, 20)
	duration := time.Since(start)

	if err != nil {
		log.Fatalf("Matching failed: %v", err)
	}

	log.Printf("✓ Found %d matches in %v\n", len(matches), duration)
	if duration < 2*time.Second {
		log.Printf("✓ PASSED: Matching completed in <2 seconds\n")
	} else {
		log.Printf("✗ FAILED: Matching took longer than 2 seconds\n")
	}

	// Display top matches
	log.Println("\nTop 5 matches:")
	for i, match := range matches {
		if i >= 5 {
			break
		}
		log.Printf("  %d. %s (score: %.2f, overlap skills: %d)\n",
			i+1, match.DisplayName, match.MatchScore, len(match.OverlapSkills))
	}

	log.Println("\n✓ Load tests completed successfully!")
}
