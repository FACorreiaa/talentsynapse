package auth

import (
	"context"
	"errors"

	"github.com/FACorreiaa/skillsphere/internal/app/user"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrPasswordMismatch   = errors.New("passwords do not match")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrInvalidEmail       = errors.New("invalid email address")
)

// ServiceInterface defines the auth service contract
type ServiceInterface interface {
	Login(ctx context.Context, input user.LoginInput) (*user.User, error)
	Register(ctx context.Context, input user.RegisterInput) (*user.User, error)
	RequestPasswordReset(ctx context.Context, email string) error
}

// Service implements authentication business logic
type Service struct {
	userRepo user.RepositoryInterface
}

// NewService creates a new auth service
func NewService(userRepo user.RepositoryInterface) *Service {
	return &Service{userRepo: userRepo}
}

// Login authenticates a user
func (s *Service) Login(ctx context.Context, input user.LoginInput) (*user.User, error) {
	// Get user by email
	u, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check password
	if !CheckPassword(u.HashedPassword, input.Password) {
		return nil, ErrInvalidCredentials
	}

	// Update last login
	_ = s.userRepo.UpdateLastLogin(ctx, u.ID)

	return u, nil
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, input user.RegisterInput) (*user.User, error) {
	// Validate input
	if !ValidateEmail(input.Email) {
		return nil, ErrInvalidEmail
	}

	if !ValidatePassword(input.Password) {
		return nil, ErrWeakPassword
	}

	if input.Password != input.ConfirmPassword {
		return nil, ErrPasswordMismatch
	}

	// Hash password
	hashedPassword, err := HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	u, err := s.userRepo.Create(ctx, input.Email, input.Username, hashedPassword, input.DisplayName)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// RequestPasswordReset initiates password reset flow
func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	// Check if user exists (but don't reveal this to caller for security)
	_, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Silently return success to prevent email enumeration
		return nil
	}

	// TODO: Generate token and send email
	// For MVP, we just return success
	return nil
}
