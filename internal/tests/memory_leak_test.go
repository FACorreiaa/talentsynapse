package tests

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// MemStats holds memory statistics for comparison
type MemStats struct {
	Alloc      uint64
	TotalAlloc uint64
	Sys        uint64
	NumGC      uint32
	Goroutines int
}

// GetMemStats returns current memory statistics
func GetMemStats() MemStats {
	runtime.GC() // Force GC to get accurate stats
	time.Sleep(50 * time.Millisecond)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemStats{
		Alloc:      m.Alloc,
		TotalAlloc: m.TotalAlloc,
		Sys:        m.Sys,
		NumGC:      m.NumGC,
		Goroutines: runtime.NumGoroutine(),
	}
}

// FormatBytes converts bytes to human-readable format
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// TestDatabaseConnectionPoolMemoryLeak tests for memory leaks in database connections
// Acceptance Criteria: No memory leaks in long-running processes
func TestDatabaseConnectionPoolMemoryLeak(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5470/myapp?sslmode=disable"
	}

	ctx := context.Background()

	// Get baseline memory stats
	baseline := GetMemStats()
	t.Logf("Baseline - Alloc: %s, Goroutines: %d", FormatBytes(baseline.Alloc), baseline.Goroutines)

	// Simulate long-running database operations
	const iterations = 100
	var pools []*pgxpool.Pool

	for i := 0; i < iterations; i++ {
		pool, err := pgxpool.New(ctx, dsn)
		if err != nil {
			t.Skipf("Could not connect to database: %v", err)
			return
		}

		// Perform some queries
		var count int
		err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE email LIKE '%@loadtest.example.com'").Scan(&count)
		require.NoError(t, err)

		pools = append(pools, pool)

		// Close every 10 connections to prevent hitting connection limits
		if i%10 == 9 {
			for _, p := range pools {
				p.Close()
			}
			pools = nil // Clear slice to allow GC
			runtime.GC()
			time.Sleep(10 * time.Millisecond)
		}
	}

	// Clean up remaining pools
	for _, p := range pools {
		p.Close()
	}
	// pools is about to go out of scope, no need to set to nil

	// Force garbage collection
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	// Get final memory stats
	final := GetMemStats()
	t.Logf("Final   - Alloc: %s, Goroutines: %d", FormatBytes(final.Alloc), final.Goroutines)

	// Calculate growth
	allocGrowth := int64(final.Alloc) - int64(baseline.Alloc)
	goroutineGrowth := final.Goroutines - baseline.Goroutines

	t.Logf("Growth  - Alloc: %s, Goroutines: %+d", FormatBytes(uint64(allocGrowth)), goroutineGrowth)

	// Memory should not grow significantly (allow 10MB growth)
	require.Less(t, allocGrowth, int64(10*1024*1024), "Memory leak detected: allocation grew by %s", FormatBytes(uint64(allocGrowth)))

	// Goroutines should not leak (allow small variance)
	require.Less(t, goroutineGrowth, 10, "Goroutine leak detected: %d new goroutines", goroutineGrowth)
}

// TestLongRunningQueryMemoryLeak tests for memory leaks in long-running queries
func TestLongRunningQueryMemoryLeak(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()

	baseline := GetMemStats()
	t.Logf("Baseline - Alloc: %s, Goroutines: %d", FormatBytes(baseline.Alloc), baseline.Goroutines)

	// Run many queries in sequence
	const iterations = 500

	for i := 0; i < iterations; i++ {
		rows, err := pool.Query(ctx, `
			SELECT u.id, u.username, u.display_name, u.email
			FROM users u
			WHERE u.email LIKE '%@loadtest.example.com'
			LIMIT 50
		`)
		require.NoError(t, err)

		// Properly consume and close rows
		count := 0
		for rows.Next() {
			var id, username, displayName, email string
			err := rows.Scan(&id, &username, &displayName, &email)
			require.NoError(t, err)
			count++
		}
		rows.Close()

		// Periodic GC
		if i%100 == 99 {
			runtime.GC()
			time.Sleep(10 * time.Millisecond)
		}
	}

	// Final cleanup
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	final := GetMemStats()
	t.Logf("Final   - Alloc: %s, Goroutines: %d", FormatBytes(final.Alloc), final.Goroutines)

	allocGrowth := int64(final.Alloc) - int64(baseline.Alloc)
	goroutineGrowth := final.Goroutines - baseline.Goroutines

	t.Logf("Growth  - Alloc: %s, Goroutines: %+d", FormatBytes(uint64(allocGrowth)), goroutineGrowth)

	// Memory should remain stable
	require.Less(t, allocGrowth, int64(5*1024*1024), "Memory leak in query execution")
	require.Less(t, goroutineGrowth, 5, "Goroutine leak in query execution")
}

// TestConcurrentOperationsMemoryLeak tests for memory leaks under concurrent load
func TestConcurrentOperationsMemoryLeak(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()

	baseline := GetMemStats()
	t.Logf("Baseline - Alloc: %s, Goroutines: %d", FormatBytes(baseline.Alloc), baseline.Goroutines)

	// Run concurrent operations
	const numWorkers = 20
	const operationsPerWorker = 50

	done := make(chan bool, numWorkers)

	for w := 0; w < numWorkers; w++ {
		go func(workerID int) {
			for i := 0; i < operationsPerWorker; i++ {
				// Query 1: User lookup
				var userID string
				_ = pool.QueryRow(ctx, `
					SELECT id FROM users WHERE email LIKE '%@loadtest.example.com' ORDER BY RANDOM() LIMIT 1
				`).Scan(&userID)

				// Query 2: Get user skills
				rows, _ := pool.Query(ctx, `
					SELECT skill_id FROM user_skills WHERE user_id = $1 LIMIT 10
				`, userID)
				if rows != nil {
					for rows.Next() {
						var skillID string
						_ = rows.Scan(&skillID)
					}
					rows.Close()
				}

				// Query 3: Get sessions
				rows2, _ := pool.Query(ctx, `
					SELECT id FROM sessions WHERE initiator_id = $1 OR partner_id = $1 LIMIT 5
				`, userID)
				if rows2 != nil {
					for rows2.Next() {
						var sessionID string
						_ = rows2.Scan(&sessionID)
					}
					rows2.Close()
				}
			}
			done <- true
		}(w)
	}

	// Wait for all workers
	for i := 0; i < numWorkers; i++ {
		<-done
	}

	// Allow time for cleanup
	runtime.GC()
	time.Sleep(200 * time.Millisecond)

	final := GetMemStats()
	t.Logf("Final   - Alloc: %s, Goroutines: %d", FormatBytes(final.Alloc), final.Goroutines)

	allocGrowth := int64(final.Alloc) - int64(baseline.Alloc)
	goroutineGrowth := final.Goroutines - baseline.Goroutines

	t.Logf("Growth  - Alloc: %s, Goroutines: %+d", FormatBytes(uint64(allocGrowth)), goroutineGrowth)

	// Under concurrent load, allow slightly more growth
	require.Less(t, allocGrowth, int64(15*1024*1024), "Memory leak under concurrent load")
	require.Less(t, goroutineGrowth, 15, "Goroutine leak under concurrent load")
}

// TestRowsClosedProperly ensures all query rows are properly closed
func TestRowsClosedProperly(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()

	baseline := GetMemStats()

	// This test simulates a common mistake: not closing rows
	// We'll test both the correct way and verify no leaks

	const iterations = 100

	for i := 0; i < iterations; i++ {
		rows, err := pool.Query(ctx, `SELECT id FROM users WHERE email LIKE '%@loadtest.example.com' LIMIT 10`)
		require.NoError(t, err)

		// Properly consume rows
		for rows.Next() {
			var id string
			_ = rows.Scan(&id)
		}

		// IMPORTANT: Always close rows
		rows.Close()

		// Check for errors that might indicate connection issues
		err = rows.Err()
		require.NoError(t, err)

		if i%25 == 24 {
			runtime.GC()
		}
	}

	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	final := GetMemStats()

	allocGrowth := int64(final.Alloc) - int64(baseline.Alloc)
	t.Logf("Memory growth with proper row closing: %s", FormatBytes(uint64(allocGrowth)))

	require.Less(t, allocGrowth, int64(3*1024*1024), "Memory should not leak with proper row handling")
}

// TestMemoryUsageUnderSustainedLoad tests memory stability during sustained operations
func TestMemoryUsageUnderSustainedLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()

	// Track memory over time
	const checkpoints = 10
	const iterationsPerCheckpoint = 100
	memorySnapshots := make([]MemStats, checkpoints)

	for checkpoint := 0; checkpoint < checkpoints; checkpoint++ {
		// Perform operations
		for i := 0; i < iterationsPerCheckpoint; i++ {
			// Mix of different operations
			var count int
			_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE email LIKE '%@loadtest.example.com'").Scan(&count)

			rows, _ := pool.Query(ctx, "SELECT id FROM sessions LIMIT 20")
			if rows != nil {
				for rows.Next() {
					var id string
					_ = rows.Scan(&id)
				}
				rows.Close()
			}

			rows2, _ := pool.Query(ctx, "SELECT id FROM conversations LIMIT 20")
			if rows2 != nil {
				for rows2.Next() {
					var id string
					_ = rows2.Scan(&id)
				}
				rows2.Close()
			}
		}

		// Take snapshot
		runtime.GC()
		time.Sleep(50 * time.Millisecond)
		memorySnapshots[checkpoint] = GetMemStats()

		t.Logf("Checkpoint %2d - Alloc: %s, Goroutines: %d",
			checkpoint, FormatBytes(memorySnapshots[checkpoint].Alloc), memorySnapshots[checkpoint].Goroutines)
	}

	// Analyze trend
	firstHalfAvg := uint64(0)
	secondHalfAvg := uint64(0)

	for i := 0; i < checkpoints/2; i++ {
		firstHalfAvg += memorySnapshots[i].Alloc
		secondHalfAvg += memorySnapshots[i+checkpoints/2].Alloc
	}
	firstHalfAvg /= uint64(checkpoints / 2)
	secondHalfAvg /= uint64(checkpoints / 2)

	growth := int64(secondHalfAvg) - int64(firstHalfAvg)
	growthPercent := float64(growth) / float64(firstHalfAvg) * 100

	t.Logf("First half average:  %s", FormatBytes(firstHalfAvg))
	t.Logf("Second half average: %s", FormatBytes(secondHalfAvg))
	t.Logf("Growth: %s (%.1f%%)", FormatBytes(uint64(growth)), growthPercent)

	// Memory should stabilize (allow 50% growth from first to second half)
	require.Less(t, growthPercent, 50.0, "Memory continues to grow - possible leak")
}

// Helper function
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
