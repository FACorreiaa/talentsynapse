package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrTokenNotFound = errors.New("token not found or expired")
	ErrTokenExpired  = errors.New("token has expired")
)

// TokenService handles token operations
type TokenService struct {
	pool *pgxpool.Pool
}

// NewTokenService creates a new token service
func NewTokenService(pool *pgxpool.Pool) *TokenService {
	return &TokenService{pool: pool}
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// HashToken creates a SHA-256 hash of the token for database storage
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}

// CreateVerificationToken generates and stores an email verification token
func (s *TokenService) CreateVerificationToken(ctx context.Context, userID string) (string, error) {
	// Generate token
	token, err := GenerateSecureToken()
	if err != nil {
		return "", err
	}

	// Hash for storage
	tokenHash := HashToken(token)

	// Store in database with 24-hour expiration
	query := `
		INSERT INTO user_tokens (token_hash, user_id, type, expires_at)
		VALUES ($1, $2, 'email_verification', NOW() + INTERVAL '24 hours')
		ON CONFLICT (token_hash) DO NOTHING
	`

	_, err = s.pool.Exec(ctx, query, tokenHash, userID)
	if err != nil {
		return "", err
	}

	return token, nil
}

// CreatePasswordResetToken generates and stores a password reset token
func (s *TokenService) CreatePasswordResetToken(ctx context.Context, userID string) (string, error) {
	// Generate token
	token, err := GenerateSecureToken()
	if err != nil {
		return "", err
	}

	// Hash for storage
	tokenHash := HashToken(token)

	// Store in database with 1-hour expiration
	query := `
		INSERT INTO user_tokens (token_hash, user_id, type, expires_at)
		VALUES ($1, $2, 'password_reset', NOW() + INTERVAL '1 hour')
		ON CONFLICT (token_hash) DO NOTHING
	`

	_, err = s.pool.Exec(ctx, query, tokenHash, userID)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ValidateVerificationToken validates an email verification token and returns the user ID
func (s *TokenService) ValidateVerificationToken(ctx context.Context, token string) (string, error) {
	tokenHash := HashToken(token)

	query := `
		SELECT user_id, expires_at
		FROM user_tokens
		WHERE token_hash = $1
		AND type = 'email_verification'
	`

	var userID string
	var expiresAt time.Time

	err := s.pool.QueryRow(ctx, query, tokenHash).Scan(&userID, &expiresAt)
	if err != nil {
		return "", ErrTokenNotFound
	}

	// Check if expired
	if time.Now().After(expiresAt) {
		// Clean up expired token
		_ = s.DeleteToken(ctx, tokenHash)
		return "", ErrTokenExpired
	}

	return userID, nil
}

// ValidatePasswordResetToken validates a password reset token and returns the user ID
func (s *TokenService) ValidatePasswordResetToken(ctx context.Context, token string) (string, error) {
	tokenHash := HashToken(token)

	query := `
		SELECT user_id, expires_at
		FROM user_tokens
		WHERE token_hash = $1
		AND type = 'password_reset'
	`

	var userID string
	var expiresAt time.Time

	err := s.pool.QueryRow(ctx, query, tokenHash).Scan(&userID, &expiresAt)
	if err != nil {
		return "", ErrTokenNotFound
	}

	// Check if expired
	if time.Now().After(expiresAt) {
		// Clean up expired token
		_ = s.DeleteToken(ctx, tokenHash)
		return "", ErrTokenExpired
	}

	return userID, nil
}

// DeleteToken removes a token from the database
func (s *TokenService) DeleteToken(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM user_tokens WHERE token_hash = $1`
	_, err := s.pool.Exec(ctx, query, tokenHash)
	return err
}

// DeleteTokenByUser removes all tokens for a specific user and type
func (s *TokenService) DeleteTokenByUser(ctx context.Context, userID, tokenType string) error {
	query := `DELETE FROM user_tokens WHERE user_id = $1 AND type = $2`
	_, err := s.pool.Exec(ctx, query, userID, tokenType)
	return err
}

// DeletePasswordResetTokens removes all password reset tokens for a user
func (s *TokenService) DeletePasswordResetTokens(ctx context.Context, userID string) error {
	return s.DeleteTokenByUser(ctx, userID, "password_reset")
}

// MarkEmailAsVerified marks a user's email as verified
func (s *TokenService) MarkEmailAsVerified(ctx context.Context, userID string) error {
	query := `
		UPDATE users
		SET email_verified_at = NOW()
		WHERE id = $1
		AND email_verified_at IS NULL
	`

	result, err := s.pool.Exec(ctx, query, userID)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		// Email was already verified or user doesn't exist
		return nil
	}

	// Clean up verification tokens for this user
	_ = s.DeleteTokenByUser(ctx, userID, "email_verification")

	return nil
}

// CleanupExpiredTokens removes all expired tokens (should be called periodically)
func (s *TokenService) CleanupExpiredTokens(ctx context.Context) error {
	query := `DELETE FROM user_tokens WHERE expires_at < NOW()`
	_, err := s.pool.Exec(ctx, query)
	return err
}
