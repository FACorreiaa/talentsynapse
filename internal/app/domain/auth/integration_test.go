package auth

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/user"
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

// createTestUser creates a test user in the database and returns the user
func createTestUser(t *testing.T, pool *pgxpool.Pool, email, username string, isActive bool) *user.User {
	ctx := context.Background()

	query := `
		INSERT INTO users (email, username, hashed_password, display_name, is_active)
		VALUES ($1, $2, '$2a$10$test', $3, $4)
		RETURNING id, email, username, display_name, role, is_active, created_at
	`

	u := &user.User{}
	err := pool.QueryRow(ctx, query, email, username, username, isActive).Scan(
		&u.ID, &u.Email, &u.Username, &u.DisplayName, &u.Role, &u.IsActive, &u.CreatedAt,
	)
	require.NoError(t, err)

	return u
}

// cleanupTestUser removes a test user from the database
func cleanupTestUser(t *testing.T, pool *pgxpool.Pool, userID string) {
	ctx := context.Background()
	_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
}

// TestIsUserActive_Integration tests IsUserActive against real database
func TestIsUserActive_Integration(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := user.NewRepository(pool)
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	tests := []struct {
		name           string
		email          string
		username       string
		isActive       bool
		expectedActive bool
	}{
		{
			name:           "active user returns true",
			email:          fmt.Sprintf("active_%s@test.com", uniqueSuffix),
			username:       fmt.Sprintf("active_%s", uniqueSuffix),
			isActive:       true,
			expectedActive: true,
		},
		{
			name:           "inactive user returns false",
			email:          fmt.Sprintf("inactive_%s@test.com", uniqueSuffix),
			username:       fmt.Sprintf("inactive_%s", uniqueSuffix),
			isActive:       false,
			expectedActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test user
			testUser := createTestUser(t, pool, tt.email, tt.username, tt.isActive)
			defer cleanupTestUser(t, pool, testUser.ID)

			// Test IsUserActive
			isActive, err := repo.IsUserActive(context.Background(), testUser.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedActive, isActive)
		})
	}
}

// TestIsUserActive_NonExistentUser_Integration tests IsUserActive for non-existent user
func TestIsUserActive_NonExistentUser_Integration(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := user.NewRepository(pool)

	// Test with non-existent user ID
	isActive, err := repo.IsUserActive(context.Background(), "00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)
	assert.False(t, isActive, "Non-existent user should return false")
}

// TestUserBanInvalidatesSession_Integration tests that banned user cannot access protected routes
func TestUserBanInvalidatesSession_Integration(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := user.NewRepository(pool)
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create an active user
	testUser := createTestUser(t, pool,
		fmt.Sprintf("ban_test_%s@test.com", uniqueSuffix),
		fmt.Sprintf("ban_test_%s", uniqueSuffix),
		true,
	)
	defer cleanupTestUser(t, pool, testUser.ID)

	// Verify user is active
	isActive, err := repo.IsUserActive(context.Background(), testUser.ID)
	require.NoError(t, err)
	assert.True(t, isActive, "User should be active initially")

	// Ban the user (set is_active = false)
	_, err = pool.Exec(context.Background(),
		"UPDATE users SET is_active = false WHERE id = $1", testUser.ID)
	require.NoError(t, err)

	// Verify user is now inactive
	isActive, err = repo.IsUserActive(context.Background(), testUser.ID)
	require.NoError(t, err)
	assert.False(t, isActive, "User should be inactive after ban")
}

// TestSoftDeletedUserCannotLogin_Integration tests that soft-deleted users are treated as inactive
func TestSoftDeletedUserCannotLogin_Integration(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := user.NewRepository(pool)
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create an active user
	testUser := createTestUser(t, pool,
		fmt.Sprintf("delete_test_%s@test.com", uniqueSuffix),
		fmt.Sprintf("delete_test_%s", uniqueSuffix),
		true,
	)
	defer cleanupTestUser(t, pool, testUser.ID)

	// Verify user is active
	isActive, err := repo.IsUserActive(context.Background(), testUser.ID)
	require.NoError(t, err)
	assert.True(t, isActive, "User should be active initially")

	// Soft delete the user
	_, err = pool.Exec(context.Background(),
		"UPDATE users SET deleted_at = NOW() WHERE id = $1", testUser.ID)
	require.NoError(t, err)

	// Verify user is now considered inactive (soft deleted)
	isActive, err = repo.IsUserActive(context.Background(), testUser.ID)
	require.NoError(t, err)
	assert.False(t, isActive, "Soft-deleted user should be treated as inactive")
}

// TestAuthenticationFlow_Integration tests the full authentication flow
func TestAuthenticationFlow_Integration(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := user.NewRepository(pool)
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("auth_flow_%s@test.com", uniqueSuffix)
	username := fmt.Sprintf("auth_flow_%s", uniqueSuffix)

	// Create user via repository
	hashedPassword := "$2a$10$test.hashed.password" // Fake hash for testing

	ctx := context.Background()

	// Insert user directly for testing
	var userID string
	err := pool.QueryRow(ctx, `
		INSERT INTO users (email, username, hashed_password, display_name, is_active)
		VALUES ($1, $2, $3, $4, true)
		RETURNING id
	`, email, username, hashedPassword, username).Scan(&userID)
	require.NoError(t, err)
	defer cleanupTestUser(t, pool, userID)

	// Test GetByEmail
	foundUser, err := repo.GetByEmail(ctx, email)
	require.NoError(t, err)
	assert.Equal(t, email, foundUser.Email)
	assert.Equal(t, username, foundUser.Username)

	// Test GetByID
	foundUser, err = repo.GetByID(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, userID, foundUser.ID)

	// Test IsUserActive
	isActive, err := repo.IsUserActive(ctx, userID)
	require.NoError(t, err)
	assert.True(t, isActive)

	// Test UpdateLastLogin
	err = repo.UpdateLastLogin(ctx, userID)
	require.NoError(t, err)

	// Verify last_login_at was updated
	var lastLogin *time.Time
	err = pool.QueryRow(ctx, "SELECT last_login_at FROM users WHERE id = $1", userID).Scan(&lastLogin)
	require.NoError(t, err)
	assert.NotNil(t, lastLogin)
}

// TestProtectedRouteAccess_Integration simulates protected route access scenarios
func TestProtectedRouteAccess_Integration(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	repo := user.NewRepository(pool)
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create different user types for testing
	activeUser := createTestUser(t, pool,
		fmt.Sprintf("active_route_%s@test.com", uniqueSuffix),
		fmt.Sprintf("active_route_%s", uniqueSuffix),
		true,
	)
	defer cleanupTestUser(t, pool, activeUser.ID)

	bannedUser := createTestUser(t, pool,
		fmt.Sprintf("banned_route_%s@test.com", uniqueSuffix),
		fmt.Sprintf("banned_route_%s", uniqueSuffix),
		false,
	)
	defer cleanupTestUser(t, pool, bannedUser.ID)

	tests := []struct {
		name         string
		userID       string
		shouldAccess bool
	}{
		{
			name:         "active user can access protected route",
			userID:       activeUser.ID,
			shouldAccess: true,
		},
		{
			name:         "banned user cannot access protected route",
			userID:       bannedUser.ID,
			shouldAccess: false,
		},
		{
			name:         "non-existent user cannot access protected route",
			userID:       "00000000-0000-0000-0000-000000000000",
			shouldAccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isActive, err := repo.IsUserActive(context.Background(), tt.userID)
			require.NoError(t, err)
			assert.Equal(t, tt.shouldAccess, isActive)
		})
	}
}

// ============================================================================
// MFA Integration Tests
// ============================================================================

// TestMFAService_Integration tests MFA service operations
func TestMFAService_Integration(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	// Create MFA service (email service can be nil for these tests)
	mfaService := NewMFAService(pool, nil)
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create a test user
	testUser := createTestUser(t, pool,
		fmt.Sprintf("mfa_test_%s@test.com", uniqueSuffix),
		fmt.Sprintf("mfa_test_%s", uniqueSuffix),
		true,
	)
	defer cleanupTestUser(t, pool, testUser.ID)

	t.Run("generate MFA setup", func(t *testing.T) {
		setup, err := mfaService.GenerateSetup(
			context.Background(),
			testUser.ID,
			testUser.Email,
			testUser.Username,
		)
		require.NoError(t, err)
		assert.NotEmpty(t, setup.Secret)
		assert.NotEmpty(t, setup.QRCodeURL)
		assert.Len(t, setup.BackupCodes, 10)

		// Verify backup codes format (XXXX-XXXX)
		for _, code := range setup.BackupCodes {
			assert.Regexp(t, `^[A-F0-9]{4}-[A-F0-9]{4}$`, code)
		}
	})

	t.Run("MFA not enabled by default after setup", func(t *testing.T) {
		enabled, err := mfaService.IsMFAEnabled(context.Background(), testUser.ID)
		require.NoError(t, err)
		assert.False(t, enabled)
	})

	t.Run("get backup codes count", func(t *testing.T) {
		count, err := mfaService.GetBackupCodesCount(context.Background(), testUser.ID)
		require.NoError(t, err)
		assert.Equal(t, 10, count)
	})
}

// TestMFARateLimit_Integration tests MFA rate limiting
func TestMFARateLimit_Integration(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	mfaService := NewMFAService(pool, nil)
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create a test user
	testUser := createTestUser(t, pool,
		fmt.Sprintf("rate_limit_%s@test.com", uniqueSuffix),
		fmt.Sprintf("rate_limit_%s", uniqueSuffix),
		true,
	)
	defer cleanupTestUser(t, pool, testUser.ID)
	defer func() {
		// Cleanup rate limit entry
		_, _ = pool.Exec(context.Background(), "DELETE FROM mfa_rate_limit WHERE user_id = $1", testUser.ID)
	}()

	t.Run("initial state not rate limited", func(t *testing.T) {
		isLocked, _, err := mfaService.CheckMFARateLimit(context.Background(), testUser.ID)
		require.NoError(t, err)
		assert.False(t, isLocked)
	})

	t.Run("record failed attempts", func(t *testing.T) {
		// Record some failed attempts
		for i := 0; i < 3; i++ {
			err := mfaService.RecordMFAAttempt(context.Background(), testUser.ID, false)
			require.NoError(t, err)
		}

		// Should not be locked yet (< MaxMFAAttempts)
		isLocked, _, err := mfaService.CheckMFARateLimit(context.Background(), testUser.ID)
		require.NoError(t, err)
		assert.False(t, isLocked)
	})

	t.Run("successful attempt resets counter", func(t *testing.T) {
		err := mfaService.RecordMFAAttempt(context.Background(), testUser.ID, true)
		require.NoError(t, err)

		isLocked, _, err := mfaService.CheckMFARateLimit(context.Background(), testUser.ID)
		require.NoError(t, err)
		assert.False(t, isLocked)
	})
}

// TestMFAAuditLog_Integration tests MFA audit logging
func TestMFAAuditLog_Integration(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	mfaService := NewMFAService(pool, nil)
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create a test user
	testUser := createTestUser(t, pool,
		fmt.Sprintf("audit_log_%s@test.com", uniqueSuffix),
		fmt.Sprintf("audit_log_%s", uniqueSuffix),
		true,
	)
	defer cleanupTestUser(t, pool, testUser.ID)
	defer func() {
		// Cleanup audit log entries
		_, _ = pool.Exec(context.Background(), "DELETE FROM mfa_audit_log WHERE user_id = $1", testUser.ID)
	}()

	t.Run("log MFA event", func(t *testing.T) {
		err := mfaService.LogMFAEvent(
			context.Background(),
			testUser.ID,
			MFAEventEnabled,
			"192.168.1.1",
			"Mozilla/5.0",
			map[string]string{"test": "value"},
		)
		require.NoError(t, err)
	})

	t.Run("retrieve audit log", func(t *testing.T) {
		entries, err := mfaService.GetMFAAuditLog(context.Background(), testUser.ID, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(entries), 1)

		// Check the logged entry
		found := false
		for _, entry := range entries {
			if entry.EventType == MFAEventEnabled {
				found = true
				assert.Equal(t, "192.168.1.1", entry.IPAddress)
				assert.Equal(t, "Mozilla/5.0", entry.UserAgent)
				break
			}
		}
		assert.True(t, found, "Should find the logged event")
	})
}

// TestTrustedDevices_Integration tests trusted device functionality
func TestTrustedDevices_Integration(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	mfaService := NewMFAService(pool, nil)
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create a test user
	testUser := createTestUser(t, pool,
		fmt.Sprintf("trusted_%s@test.com", uniqueSuffix),
		fmt.Sprintf("trusted_%s", uniqueSuffix),
		true,
	)
	defer cleanupTestUser(t, pool, testUser.ID)
	defer func() {
		// Cleanup trusted devices
		_, _ = pool.Exec(context.Background(), "DELETE FROM trusted_devices WHERE user_id = $1", testUser.ID)
		_, _ = pool.Exec(context.Background(), "DELETE FROM mfa_audit_log WHERE user_id = $1", testUser.ID)
	}()

	var deviceToken string

	t.Run("add trusted device", func(t *testing.T) {
		var err error
		deviceToken, err = mfaService.AddTrustedDevice(
			context.Background(),
			testUser.ID,
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			"192.168.1.100",
		)
		require.NoError(t, err)
		assert.NotEmpty(t, deviceToken)
	})

	t.Run("verify trusted device", func(t *testing.T) {
		isTrusted, err := mfaService.IsTrustedDevice(context.Background(), testUser.ID, deviceToken)
		require.NoError(t, err)
		assert.True(t, isTrusted)
	})

	t.Run("invalid token not trusted", func(t *testing.T) {
		isTrusted, err := mfaService.IsTrustedDevice(context.Background(), testUser.ID, "invalid-token")
		require.NoError(t, err)
		assert.False(t, isTrusted)
	})

	t.Run("empty token not trusted", func(t *testing.T) {
		isTrusted, err := mfaService.IsTrustedDevice(context.Background(), testUser.ID, "")
		require.NoError(t, err)
		assert.False(t, isTrusted)
	})

	t.Run("list trusted devices", func(t *testing.T) {
		devices, err := mfaService.GetTrustedDevices(context.Background(), testUser.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(devices), 1)
	})

	t.Run("remove all trusted devices", func(t *testing.T) {
		err := mfaService.RemoveAllTrustedDevices(context.Background(), testUser.ID)
		require.NoError(t, err)

		// Verify token no longer works
		isTrusted, err := mfaService.IsTrustedDevice(context.Background(), testUser.ID, deviceToken)
		require.NoError(t, err)
		assert.False(t, isTrusted)
	})
}
