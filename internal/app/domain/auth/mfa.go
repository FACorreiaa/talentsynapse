package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pquerna/otp/totp"
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
