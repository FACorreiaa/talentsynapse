package dashboard

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// TestDashboardLoadPerformance tests dashboard query performance
// Acceptance Criteria: Dashboard loads in <1 second
func TestDashboardLoadPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()

	// Get a test user with sessions and matches
	var testUserID string
	err := pool.QueryRow(ctx, `
		SELECT u.id FROM users u
		WHERE u.email LIKE '%@loadtest.example.com'
		AND EXISTS (SELECT 1 FROM sessions s WHERE s.initiator_id = u.id OR s.partner_id = u.id)
		ORDER BY RANDOM()
		LIMIT 1
	`).Scan(&testUserID)
	if err != nil {
		t.Skipf("Skipping test: no load test data found (run seed script first): %v", err)
		return
	}

	t.Logf("Testing dashboard performance for user: %s", testUserID)

	// Test 1: User stats query
	t.Run("UserStatsQuery", func(t *testing.T) {
		start := time.Now()

		var totalSessions, completedSessions, totalReviews int
		var avgRating float64
		err := pool.QueryRow(ctx, `
			SELECT
				total_sessions,
				completed_sessions,
				average_rating,
				total_reviews
			FROM user_stats
			WHERE user_id = $1
		`, testUserID).Scan(&totalSessions, &completedSessions, &avgRating, &totalReviews)

		duration := time.Since(start)

		require.NoError(t, err)
		t.Logf("Stats query took: %v (sessions: %d, completed: %d, rating: %.2f, reviews: %d)",
			duration, totalSessions, completedSessions, avgRating, totalReviews)

		require.Less(t, duration, 100*time.Millisecond, "Stats query should be very fast (<100ms)")
	})

	// Test 2: Upcoming sessions query
	t.Run("UpcomingSessionsQuery", func(t *testing.T) {
		start := time.Now()

		rows, err := pool.Query(ctx, `
			SELECT
				s.id,
				s.status,
				s.scheduled_start,
				s.scheduled_end,
				CASE
					WHEN s.initiator_id = $1 THEN partner.display_name
					ELSE initiator.display_name
				END as partner_name,
				CASE
					WHEN s.initiator_id = $1 THEN partner.avatar_url
					ELSE initiator.avatar_url
				END as partner_avatar
			FROM sessions s
			JOIN users initiator ON initiator.id = s.initiator_id
			JOIN users partner ON partner.id = s.partner_id
			WHERE (s.initiator_id = $1 OR s.partner_id = $1)
			AND s.status IN ('pending', 'confirmed', 'in_progress')
			AND s.scheduled_start > NOW()
			ORDER BY s.scheduled_start ASC
			LIMIT 10
		`, testUserID)

		duration := time.Since(start)

		require.NoError(t, err)
		defer rows.Close()

		sessionCount := 0
		for rows.Next() {
			sessionCount++
			var id, status, partnerName string
			var scheduledStart, scheduledEnd time.Time
			var partnerAvatar *string
			err := rows.Scan(&id, &status, &scheduledStart, &scheduledEnd, &partnerName, &partnerAvatar)
			require.NoError(t, err)
		}

		t.Logf("Upcoming sessions query took: %v (found %d sessions)", duration, sessionCount)
		require.Less(t, duration, 200*time.Millisecond, "Upcoming sessions query should be fast (<200ms)")
	})

	// Test 3: Recent sessions query
	t.Run("RecentSessionsQuery", func(t *testing.T) {
		start := time.Now()

		rows, err := pool.Query(ctx, `
			SELECT
				s.id,
				s.status,
				s.scheduled_start,
				s.scheduled_end,
				CASE
					WHEN s.initiator_id = $1 THEN partner.display_name
					ELSE initiator.display_name
				END as partner_name,
				CASE
					WHEN s.initiator_id = $1 THEN partner.avatar_url
					ELSE initiator.avatar_url
				END as partner_avatar
			FROM sessions s
			JOIN users initiator ON initiator.id = s.initiator_id
			JOIN users partner ON partner.id = s.partner_id
			WHERE (s.initiator_id = $1 OR s.partner_id = $1)
			AND s.status = 'completed'
			ORDER BY s.scheduled_start DESC
			LIMIT 10
		`, testUserID)

		duration := time.Since(start)

		require.NoError(t, err)
		defer rows.Close()

		sessionCount := 0
		for rows.Next() {
			sessionCount++
			var id, status, partnerName string
			var scheduledStart, scheduledEnd time.Time
			var partnerAvatar *string
			err := rows.Scan(&id, &status, &scheduledStart, &scheduledEnd, &partnerName, &partnerAvatar)
			require.NoError(t, err)
		}

		t.Logf("Recent sessions query took: %v (found %d sessions)", duration, sessionCount)
		require.Less(t, duration, 200*time.Millisecond, "Recent sessions query should be fast (<200ms)")
	})

	// Test 4: User badges query
	t.Run("UserBadgesQuery", func(t *testing.T) {
		start := time.Now()

		rows, err := pool.Query(ctx, `
			SELECT
				b.code,
				b.name,
				b.description,
				b.icon_url,
				ub.awarded_at
			FROM user_badges ub
			JOIN badges b ON b.id = ub.badge_id
			WHERE ub.user_id = $1
			ORDER BY ub.awarded_at DESC
		`, testUserID)

		duration := time.Since(start)

		require.NoError(t, err)
		defer rows.Close()

		badgeCount := 0
		for rows.Next() {
			badgeCount++
		}

		t.Logf("User badges query took: %v (found %d badges)", duration, badgeCount)
		require.Less(t, duration, 100*time.Millisecond, "Badges query should be very fast (<100ms)")
	})

	// Test 5: Combined dashboard load (simulates full page load)
	t.Run("FullDashboardLoad", func(t *testing.T) {
		start := time.Now()

		// Simulate parallel queries that would load on dashboard
		type result struct {
			name     string
			duration time.Duration
			err      error
		}

		results := make(chan result, 4)

		// Query 1: User stats
		go func() {
			queryStart := time.Now()
			var totalSessions, completedSessions, totalReviews int
			var avgRating float64
			err := pool.QueryRow(ctx, `
				SELECT total_sessions, completed_sessions, average_rating, total_reviews
				FROM user_stats WHERE user_id = $1
			`, testUserID).Scan(&totalSessions, &completedSessions, &avgRating, &totalReviews)
			results <- result{"UserStats", time.Since(queryStart), err}
		}()

		// Query 2: Upcoming sessions
		go func() {
			queryStart := time.Now()
			rows, err := pool.Query(ctx, `
				SELECT s.id, s.status, s.scheduled_start
				FROM sessions s
				WHERE (s.initiator_id = $1 OR s.partner_id = $1)
				AND s.status IN ('pending', 'confirmed')
				AND s.scheduled_start > NOW()
				ORDER BY s.scheduled_start ASC
				LIMIT 5
			`, testUserID)
			if err == nil {
				rows.Close()
			}
			results <- result{"UpcomingSessions", time.Since(queryStart), err}
		}()

		// Query 3: Recent activity
		go func() {
			queryStart := time.Now()
			rows, err := pool.Query(ctx, `
				SELECT s.id, s.status, s.scheduled_start
				FROM sessions s
				WHERE (s.initiator_id = $1 OR s.partner_id = $1)
				ORDER BY s.created_at DESC
				LIMIT 5
			`, testUserID)
			if err == nil {
				rows.Close()
			}
			results <- result{"RecentActivity", time.Since(queryStart), err}
		}()

		// Query 4: User badges
		go func() {
			queryStart := time.Now()
			rows, err := pool.Query(ctx, `
				SELECT ub.badge_id FROM user_badges ub WHERE ub.user_id = $1
			`, testUserID)
			if err == nil {
				rows.Close()
			}
			results <- result{"UserBadges", time.Since(queryStart), err}
		}()

		// Collect results
		for i := 0; i < 4; i++ {
			res := <-results
			require.NoError(t, res.err, fmt.Sprintf("%s query failed", res.name))
			t.Logf("  %s: %v", res.name, res.duration)
		}

		totalDuration := time.Since(start)
		t.Logf("Full dashboard load took: %v", totalDuration)
		require.Less(t, totalDuration, 1*time.Second, "Full dashboard should load in <1 second")
	})
}

// TestMatchesPageLoadPerformance tests the matches/discovery page performance
func TestMatchesPageLoadPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()

	var testUserID string
	err := pool.QueryRow(ctx, `
		SELECT id FROM users WHERE email LIKE '%@loadtest.example.com' ORDER BY RANDOM() LIMIT 1
	`).Scan(&testUserID)
	require.NoError(t, err)

	t.Run("MatchesPageQueries", func(t *testing.T) {
		start := time.Now()

		// Simulate matches page queries running in parallel
		done := make(chan bool, 3)
		errors := make(chan error, 3)

		// Query 1: Get matches
		go func() {
			queryStart := time.Now()
			rows, err := pool.Query(ctx, `
				SELECT
					u.id,
					u.display_name,
					u.username,
					u.avatar_url
				FROM user_skill_vectors target_user
				JOIN users u ON u.id = target_user.user_id
				CROSS JOIN user_skill_vectors query_user
				WHERE query_user.user_id = $1
				AND target_user.user_id != $1
				AND target_user.user_id NOT IN (
					SELECT user_id_b FROM match_history WHERE user_id_a = $1
				)
				LIMIT 20
			`, testUserID)

			if err != nil {
				errors <- err
			} else {
				rows.Close()
				t.Logf("  Matches query: %v", time.Since(queryStart))
				done <- true
			}
		}()

		// Query 2: Get user's skills
		go func() {
			queryStart := time.Now()
			rows, err := pool.Query(ctx, `
				SELECT s.name, us.skill_type, us.proficiency
				FROM user_skills us
				JOIN skills s ON s.id = us.skill_id
				WHERE us.user_id = $1
				ORDER BY us.skill_type, s.name
			`, testUserID)

			if err != nil {
				errors <- err
			} else {
				rows.Close()
				t.Logf("  User skills query: %v", time.Since(queryStart))
				done <- true
			}
		}()

		// Query 3: Get match history stats
		go func() {
			queryStart := time.Now()
			var matchCount int
			err := pool.QueryRow(ctx, `
				SELECT COUNT(*) FROM match_history WHERE user_id_a = $1
			`, testUserID).Scan(&matchCount)

			if err != nil {
				errors <- err
			} else {
				t.Logf("  Match history query: %v (count: %d)", time.Since(queryStart), matchCount)
				done <- true
			}
		}()

		// Wait for completion
		successCount := 0
		for i := 0; i < 3; i++ {
			select {
			case <-done:
				successCount++
			case err := <-errors:
				t.Errorf("Query error: %v", err)
			case <-time.After(5 * time.Second):
				t.Error("Query timeout")
			}
		}

		totalDuration := time.Since(start)
		t.Logf("Matches page load took: %v", totalDuration)

		require.Equal(t, 3, successCount, "All queries should succeed")
		require.Less(t, totalDuration, 1*time.Second, "Matches page should load in <1 second")
	})
}

// BenchmarkDashboardQueries benchmarks dashboard query performance
func BenchmarkDashboardQueries(b *testing.B) {
	pool := getBenchDBPool(b)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()

	var testUserID string
	err := pool.QueryRow(ctx, `
		SELECT id FROM users WHERE email LIKE '%@loadtest.example.com' ORDER BY RANDOM() LIMIT 1
	`).Scan(&testUserID)
	if err != nil {
		b.Skipf("No test data: %v", err)
		return
	}

	b.Run("UserStats", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var total, completed, reviews int
			var rating float64
			_ = pool.QueryRow(ctx, `
				SELECT total_sessions, completed_sessions, average_rating, total_reviews
				FROM user_stats WHERE user_id = $1
			`, testUserID).Scan(&total, &completed, &rating, &reviews)
		}
	})

	b.Run("UpcomingSessions", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			rows, _ := pool.Query(ctx, `
				SELECT s.id FROM sessions s
				WHERE (s.initiator_id = $1 OR s.partner_id = $1)
				AND s.status IN ('pending', 'confirmed')
				AND s.scheduled_start > NOW()
				LIMIT 5
			`, testUserID)
			if rows != nil {
				rows.Close()
			}
		}
	})
}

// Helper functions
func getTestDBPool(t *testing.T) *pgxpool.Pool {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5470/myapp?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("Skipping test: could not connect to database: %v", err)
		return nil
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("Skipping test: database not responding: %v", err)
		return nil
	}

	return pool
}

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
