package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidToken       = errors.New("invalid or expired reset token")
)

// Repository defines the interface for user data access
type RepositoryInterface interface {
	Create(ctx context.Context, email, username, hashedPassword, displayName string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	UpdateLastLogin(ctx context.Context, userID string) error
}

// Repository implements user data access
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new user repository
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new user into the database
func (r *Repository) Create(ctx context.Context, email, username, hashedPassword, displayName string) (*User, error) {
	query := `
		INSERT INTO users (email, username, hashed_password, display_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, username, hashed_password, display_name, avatar_url, role, is_active, email_verified_at, created_at, updated_at, last_login_at
	`

	user := &User{}
	err := r.pool.QueryRow(ctx, query, email, username, hashedPassword, displayName).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.HashedPassword,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Role,
		&user.IsActive,
		&user.EmailVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if isUniqueViolation(err, "users_email_key") {
			return nil, ErrEmailAlreadyExists
		}
		if isUniqueViolation(err, "users_username_key") {
			return nil, ErrUsernameExists
		}
		return nil, err
	}

	return user, nil
}

// GetByEmail retrieves a user by their email address
func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, username, hashed_password, display_name, avatar_url, role, is_active, email_verified_at, created_at, updated_at, last_login_at
		FROM users
		WHERE email = $1 AND is_active = true
	`

	user := &User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.HashedPassword,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Role,
		&user.IsActive,
		&user.EmailVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// GetByID retrieves a user by their ID
func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, email, username, hashed_password, display_name, avatar_url, role, is_active, email_verified_at, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1 AND is_active = true
	`

	user := &User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.HashedPassword,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Role,
		&user.IsActive,
		&user.EmailVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// GetByUsername retrieves a user by their username
func (r *Repository) GetByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT id, email, username, hashed_password, display_name, avatar_url, role, is_active, email_verified_at, created_at, updated_at, last_login_at
		FROM users
		WHERE username = $1 AND is_active = true
	`

	user := &User{}
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.HashedPassword,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Role,
		&user.IsActive,
		&user.EmailVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// UpdateLastLogin updates the user's last login timestamp
func (r *Repository) UpdateLastLogin(ctx context.Context, userID string) error {
	query := `UPDATE users SET last_login_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

// isUniqueViolation checks if error is a unique constraint violation
func isUniqueViolation(err error, constraintName string) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "23505") && contains(errStr, constraintName)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
