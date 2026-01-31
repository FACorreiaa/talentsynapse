package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pquerna/otp/totp"
)

// MFA constants
const (
	TrustedDeviceCookieName = "trusted_device"
	TrustedDeviceDuration   = 30 * 24 * time.Hour // 30 days
	MaxMFAAttempts          = 5
	MFALockoutDuration      = 15 * time.Minute
)

// MFA event types for audit logging
const (
	MFAEventEnabled                = "mfa_enabled"
	MFAEventDisabled               = "mfa_disabled"
	MFAEventVerifySuccess          = "mfa_verify_success"
	MFAEventVerifyFailed           = "mfa_verify_failed"
	MFAEventBackupCodeUsed         = "backup_code_used"
	MFAEventTrustedDeviceAdded     = "trusted_device_added"
	MFAEventTrustedDeviceRemoved   = "trusted_device_removed"
	MFAEventBackupCodesRegenerated = "backup_codes_regenerated"
)

// MFAService handles MFA operations
type MFAService struct {
	pool         *pgxpool.Pool
	emailService *EmailService
}

// NewMFAService creates a new MFA service
func NewMFAService(pool *pgxpool.Pool, emailService *EmailService) *MFAService {
	return &MFAService{
		pool:         pool,
		emailService: emailService,
	}
}

// MFASetup contains data needed to set up MFA
type MFASetup struct {
	Secret      string
	QRCodeURL   string
	BackupCodes []string
}

// GenerateSetup creates a new TOTP secret and backup codes
func (s *MFAService) GenerateSetup(ctx context.Context, userID, email, username string) (*MFASetup, error) {
	// Generate TOTP secret
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "TalentSynapse",
		AccountName: email,
		SecretSize:  32,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	// Generate backup codes
	backupCodes, err := s.generateBackupCodes(10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Hash backup codes for storage
	hashedCodes := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		hashedCodes[i] = hashBackupCode(code)
	}

	// Store in database (disabled until user confirms)
	query := `
		INSERT INTO user_mfa (user_id, secret, enabled, backup_codes)
		VALUES ($1, $2, false, $3)
		ON CONFLICT (user_id)
		DO UPDATE SET
			secret = EXCLUDED.secret,
			backup_codes = EXCLUDED.backup_codes,
			updated_at = NOW()
	`
	_, err = s.pool.Exec(ctx, query, userID, key.Secret(), hashedCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to store MFA setup: %w", err)
	}

	return &MFASetup{
		Secret:      key.Secret(),
		QRCodeURL:   key.URL(),
		BackupCodes: backupCodes,
	}, nil
}

// EnableMFA enables MFA after successful verification
func (s *MFAService) EnableMFA(ctx context.Context, userID, code string) error {
	// Get secret from database
	var secret string
	query := `SELECT secret FROM user_mfa WHERE user_id = $1`
	err := s.pool.QueryRow(ctx, query, userID).Scan(&secret)
	if err != nil {
		return fmt.Errorf("MFA not set up: %w", err)
	}

	// Verify the code
	valid := totp.Validate(code, secret)
	if !valid {
		return fmt.Errorf("invalid verification code")
	}

	// Enable MFA
	updateQuery := `
		UPDATE user_mfa
		SET enabled = true, updated_at = NOW()
		WHERE user_id = $1
	`
	_, err = s.pool.Exec(ctx, updateQuery, userID)
	if err != nil {
		return fmt.Errorf("failed to enable MFA: %w", err)
	}

	// Get user info for email
	var email, username string
	userQuery := `SELECT email, display_name FROM users WHERE id = $1`
	err = s.pool.QueryRow(ctx, userQuery, userID).Scan(&email, &username)
	if err == nil {
		// Send notification email (don't fail if email fails)
		_ = s.emailService.SendMFAEmail(email, username, true)
	}

	return nil
}

// DisableMFA disables MFA for a user
func (s *MFAService) DisableMFA(ctx context.Context, userID string) error {
	// Get user info for email (don't fail if this fails)
	var email, username string
	userQuery := `SELECT email, display_name FROM users WHERE id = $1`
	_ = s.pool.QueryRow(ctx, userQuery, userID).Scan(&email, &username)

	// Disable MFA
	query := `
		UPDATE user_mfa
		SET enabled = false, updated_at = NOW()
		WHERE user_id = $1
	`
	_, err := s.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to disable MFA: %w", err)
	}

	// Send notification email (don't fail if email fails)
	if email != "" {
		_ = s.emailService.SendMFAEmail(email, username, false)
	}

	return nil
}

// VerifyTOTP verifies a TOTP code
func (s *MFAService) VerifyTOTP(ctx context.Context, userID, code string) (bool, error) {
	var secret string
	var enabled bool
	query := `SELECT secret, enabled FROM user_mfa WHERE user_id = $1`
	err := s.pool.QueryRow(ctx, query, userID).Scan(&secret, &enabled)
	if err != nil {
		return false, fmt.Errorf("MFA not configured: %w", err)
	}

	if !enabled {
		return false, fmt.Errorf("MFA is not enabled")
	}

	// Verify code
	valid := totp.Validate(code, secret)
	if valid {
		// Update last_used_at
		updateQuery := `UPDATE user_mfa SET last_used_at = NOW() WHERE user_id = $1`
		_, _ = s.pool.Exec(ctx, updateQuery, userID)
	}

	return valid, nil
}

// VerifyBackupCode verifies and consumes a backup code
func (s *MFAService) VerifyBackupCode(ctx context.Context, userID, code string) (bool, error) {
	var backupCodes []string
	var enabled bool
	query := `SELECT backup_codes, enabled FROM user_mfa WHERE user_id = $1`
	err := s.pool.QueryRow(ctx, query, userID).Scan(&backupCodes, &enabled)
	if err != nil {
		return false, fmt.Errorf("MFA not configured: %w", err)
	}

	if !enabled {
		return false, fmt.Errorf("MFA is not enabled")
	}

	// Hash the provided code
	hashedCode := hashBackupCode(code)

	// Check if code exists
	found := false
	newCodes := make([]string, 0, len(backupCodes))
	for _, stored := range backupCodes {
		if stored == hashedCode {
			found = true
			// Don't add to newCodes (consume it)
		} else {
			newCodes = append(newCodes, stored)
		}
	}

	if !found {
		return false, nil
	}

	// Update backup codes (remove used one)
	updateQuery := `
		UPDATE user_mfa
		SET backup_codes = $1, last_used_at = NOW(), updated_at = NOW()
		WHERE user_id = $2
	`
	_, err = s.pool.Exec(ctx, updateQuery, newCodes, userID)
	if err != nil {
		return false, fmt.Errorf("failed to update backup codes: %w", err)
	}

	return true, nil
}

// IsMFAEnabled checks if MFA is enabled for a user
func (s *MFAService) IsMFAEnabled(ctx context.Context, userID string) (bool, error) {
	var enabled bool
	query := `SELECT enabled FROM user_mfa WHERE user_id = $1`
	err := s.pool.QueryRow(ctx, query, userID).Scan(&enabled)
	if err != nil {
		// If no row found, MFA is not enabled
		return false, nil
	}
	return enabled, nil
}

// GetBackupCodesCount returns the number of remaining backup codes
func (s *MFAService) GetBackupCodesCount(ctx context.Context, userID string) (int, error) {
	var backupCodes []string
	query := `SELECT backup_codes FROM user_mfa WHERE user_id = $1`
	err := s.pool.QueryRow(ctx, query, userID).Scan(&backupCodes)
	if err != nil {
		return 0, fmt.Errorf("MFA not configured: %w", err)
	}
	return len(backupCodes), nil
}

// RegenerateBackupCodes generates new backup codes
func (s *MFAService) RegenerateBackupCodes(ctx context.Context, userID string) ([]string, error) {
	// Generate new backup codes
	backupCodes, err := s.generateBackupCodes(10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Hash for storage
	hashedCodes := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		hashedCodes[i] = hashBackupCode(code)
	}

	// Update database
	query := `
		UPDATE user_mfa
		SET backup_codes = $1, updated_at = NOW()
		WHERE user_id = $2
	`
	_, err = s.pool.Exec(ctx, query, hashedCodes, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update backup codes: %w", err)
	}

	return backupCodes, nil
}

// generateBackupCodes generates cryptographically secure backup codes
func (s *MFAService) generateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		// Generate 8 bytes of random data
		b := make([]byte, 8)
		if _, err := rand.Read(b); err != nil {
			return nil, err
		}
		// Format as XXXX-XXXX
		code := fmt.Sprintf("%04X-%04X",
			uint16(b[0])<<8|uint16(b[1]),
			uint16(b[2])<<8|uint16(b[3]),
		)
		codes[i] = code
	}
	return codes, nil
}

// hashBackupCode hashes a backup code for secure storage
func hashBackupCode(code string) string {
	hash := sha256.Sum256([]byte(code))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// ============================================================================
// Trusted Device Management
// ============================================================================

// TrustedDevice represents a trusted device
type TrustedDevice struct {
	ID         string
	DeviceName string
	LastUsedAt time.Time
	CreatedAt  time.Time
}

// AddTrustedDevice adds a new trusted device for the user
// Returns empty string if table doesn't exist (graceful degradation)
func (s *MFAService) AddTrustedDevice(ctx context.Context, userID, userAgent, ipAddress string) (string, error) {
	// Generate a secure random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate device token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)
	tokenHash := hashDeviceToken(token)

	// Extract device name from user agent
	deviceName := parseDeviceName(userAgent)

	// Insert trusted device
	query := `
		INSERT INTO trusted_devices (user_id, device_token_hash, device_name, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	expiresAt := time.Now().Add(TrustedDeviceDuration)
	_, err := s.pool.Exec(ctx, query, userID, tokenHash, deviceName, userAgent, ipAddress, expiresAt)
	if err != nil {
		// Table may not exist - return empty token (graceful degradation)
		return "", nil
	}

	// Log the event (will silently fail if audit table doesn't exist)
	_ = s.LogMFAEvent(ctx, userID, MFAEventTrustedDeviceAdded, ipAddress, userAgent, map[string]string{
		"device_name": deviceName,
	})

	return token, nil
}

// IsTrustedDevice checks if the device is trusted
// Returns false if table doesn't exist (graceful degradation)
func (s *MFAService) IsTrustedDevice(ctx context.Context, userID, deviceToken string) (bool, error) {
	if deviceToken == "" {
		return false, nil
	}

	tokenHash := hashDeviceToken(deviceToken)

	query := `
		SELECT EXISTS(
			SELECT 1 FROM trusted_devices 
			WHERE user_id = $1 AND device_token_hash = $2 AND expires_at > NOW()
		)
	`
	var exists bool
	err := s.pool.QueryRow(ctx, query, userID, tokenHash).Scan(&exists)
	if err != nil {
		// Table may not exist - device is not trusted
		return false, nil
	}

	if exists {
		// Update last_used_at
		updateQuery := `UPDATE trusted_devices SET last_used_at = NOW() WHERE device_token_hash = $1`
		_, _ = s.pool.Exec(ctx, updateQuery, tokenHash)
	}

	return exists, nil
}

// GetTrustedDevices returns all trusted devices for a user
// Returns empty slice if table doesn't exist (graceful degradation)
func (s *MFAService) GetTrustedDevices(ctx context.Context, userID string) ([]TrustedDevice, error) {
	query := `
		SELECT id, device_name, last_used_at, created_at
		FROM trusted_devices
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY last_used_at DESC
	`
	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		// Table may not exist - return empty slice
		return []TrustedDevice{}, nil
	}
	defer rows.Close()

	var devices []TrustedDevice
	for rows.Next() {
		var d TrustedDevice
		if err := rows.Scan(&d.ID, &d.DeviceName, &d.LastUsedAt, &d.CreatedAt); err != nil {
			continue
		}
		devices = append(devices, d)
	}
	return devices, nil
}

// RemoveTrustedDevice removes a trusted device
// Silently fails if table doesn't exist (graceful degradation)
func (s *MFAService) RemoveTrustedDevice(ctx context.Context, userID, deviceID string) error {
	query := `DELETE FROM trusted_devices WHERE id = $1 AND user_id = $2`
	_, _ = s.pool.Exec(ctx, query, deviceID, userID) // Ignore errors - table may not exist

	_ = s.LogMFAEvent(ctx, userID, MFAEventTrustedDeviceRemoved, "", "", map[string]string{
		"device_id": deviceID,
	})

	return nil
}

// RemoveAllTrustedDevices removes all trusted devices for a user
// Silently fails if table doesn't exist (graceful degradation)
func (s *MFAService) RemoveAllTrustedDevices(ctx context.Context, userID string) error {
	query := `DELETE FROM trusted_devices WHERE user_id = $1`
	_, _ = s.pool.Exec(ctx, query, userID) // Ignore errors - table may not exist
	return nil
}

// SetTrustedDeviceCookie sets the trusted device cookie
func SetTrustedDeviceCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     TrustedDeviceCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(TrustedDeviceDuration.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// GetTrustedDeviceToken gets the trusted device token from cookie
func GetTrustedDeviceToken(r *http.Request) string {
	cookie, err := r.Cookie(TrustedDeviceCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// ClearTrustedDeviceCookie clears the trusted device cookie
func ClearTrustedDeviceCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     TrustedDeviceCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// hashDeviceToken hashes a device token for storage
func hashDeviceToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// parseDeviceName extracts a user-friendly device name from user agent
func parseDeviceName(userAgent string) string {
	// Simple parsing - could be improved with a proper user agent parser
	if len(userAgent) > 100 {
		userAgent = userAgent[:100]
	}
	if userAgent == "" {
		return "Unknown Device"
	}
	return userAgent
}

// ============================================================================
// Rate Limiting
// ============================================================================

// CheckMFARateLimit checks if the user is rate limited
// Returns false if the table doesn't exist (graceful degradation)
func (s *MFAService) CheckMFARateLimit(ctx context.Context, userID string) (bool, time.Time, error) {
	query := `
		SELECT attempt_count, locked_until
		FROM mfa_rate_limit
		WHERE user_id = $1
	`
	var attemptCount int
	var lockedUntil *time.Time
	err := s.pool.QueryRow(ctx, query, userID).Scan(&attemptCount, &lockedUntil)
	if err != nil {
		// No rate limit entry exists or table doesn't exist - user is not rate limited
		// Graceful degradation: if table missing, just allow the request
		return false, time.Time{}, nil
	}

	if lockedUntil != nil && lockedUntil.After(time.Now()) {
		return true, *lockedUntil, nil
	}

	return false, time.Time{}, nil
}

// RecordMFAAttempt records an MFA verification attempt
// Silently fails if table doesn't exist (graceful degradation)
func (s *MFAService) RecordMFAAttempt(ctx context.Context, userID string, success bool) error {
	if success {
		// Reset rate limit on success
		query := `
			INSERT INTO mfa_rate_limit (user_id, attempt_count, last_attempt_at, locked_until)
			VALUES ($1, 0, NOW(), NULL)
			ON CONFLICT (user_id) DO UPDATE SET
				attempt_count = 0,
				last_attempt_at = NOW(),
				locked_until = NULL
		`
		_, _ = s.pool.Exec(ctx, query, userID) // Ignore errors - table may not exist
		return nil
	}

	// Increment attempt count on failure
	query := `
		INSERT INTO mfa_rate_limit (user_id, attempt_count, last_attempt_at)
		VALUES ($1, 1, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			attempt_count = mfa_rate_limit.attempt_count + 1,
			last_attempt_at = NOW(),
			locked_until = CASE 
				WHEN mfa_rate_limit.attempt_count + 1 >= $2 THEN NOW() + $3::interval
				ELSE mfa_rate_limit.locked_until
			END
	`
	_, _ = s.pool.Exec(ctx, query, userID, MaxMFAAttempts, fmt.Sprintf("%d minutes", int(MFALockoutDuration.Minutes()))) // Ignore errors
	return nil
}

// ============================================================================
// Audit Logging
// ============================================================================

// LogMFAEvent logs an MFA event for security auditing
// Silently fails if table doesn't exist (graceful degradation)
func (s *MFAService) LogMFAEvent(ctx context.Context, userID, eventType, ipAddress, userAgent string, details map[string]string) error {
	var detailsJSON []byte
	var err error
	if details != nil {
		detailsJSON, err = json.Marshal(details)
		if err != nil {
			detailsJSON = []byte("{}")
		}
	} else {
		detailsJSON = []byte("{}")
	}

	query := `
		INSERT INTO mfa_audit_log (user_id, event_type, ip_address, user_agent, details)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, _ = s.pool.Exec(ctx, query, userID, eventType, ipAddress, userAgent, detailsJSON) // Ignore errors - table may not exist
	return nil
}

// GetMFAAuditLog retrieves MFA audit log entries for a user
// Returns empty slice if table doesn't exist (graceful degradation)
func (s *MFAService) GetMFAAuditLog(ctx context.Context, userID string, limit int) ([]MFAAuditEntry, error) {
	query := `
		SELECT event_type, ip_address, user_agent, details, created_at
		FROM mfa_audit_log
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := s.pool.Query(ctx, query, userID, limit)
	if err != nil {
		// Table may not exist - return empty slice
		return []MFAAuditEntry{}, nil
	}
	defer rows.Close()

	var entries []MFAAuditEntry
	for rows.Next() {
		var e MFAAuditEntry
		var detailsJSON []byte
		if err := rows.Scan(&e.EventType, &e.IPAddress, &e.UserAgent, &detailsJSON, &e.CreatedAt); err != nil {
			continue
		}
		if len(detailsJSON) > 0 {
			_ = json.Unmarshal(detailsJSON, &e.Details)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// MFAAuditEntry represents an audit log entry
type MFAAuditEntry struct {
	EventType string
	IPAddress string
	UserAgent string
	Details   map[string]string
	CreatedAt time.Time
}
