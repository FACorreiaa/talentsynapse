package auth

import (
	"context"
	"errors"

	"github.com/FACorreiaa/talentsynapse/internal/app/domain/user"
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
	ResetPassword(ctx context.Context, token, newPassword string) error
	VerifyEmail(ctx context.Context, token string) (string, error)
	ResendVerificationEmail(ctx context.Context, email string) error
}

// Service implements authentication business logic
type Service struct {
	userRepo     user.RepositoryInterface
	tokenService *TokenService
	emailService *EmailService
	mfaService   *MFAService
}

// NewService creates a new auth service
func NewService(userRepo user.RepositoryInterface, tokenService *TokenService, emailService *EmailService, mfaService *MFAService) *Service {
	return &Service{
		userRepo:     userRepo,
		tokenService: tokenService,
		emailService: emailService,
		mfaService:   mfaService,
	}
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

	// Generate verification token
	if s.tokenService != nil && s.emailService != nil {
		token, err := s.tokenService.CreateVerificationToken(ctx, u.ID)
		if err == nil {
			// Send verification email (don't fail registration if email fails)
			_ = s.emailService.SendVerificationEmail(u.Email, u.DisplayName, token)
		}
	}

	return u, nil
}

// RequestPasswordReset initiates password reset flow
func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	// Check if user exists (but don't reveal this to caller for security)
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Silently return success to prevent email enumeration
		return nil
	}

	// Generate password reset token
	if s.tokenService != nil && s.emailService != nil {
		token, err := s.tokenService.CreatePasswordResetToken(ctx, u.ID)
		if err == nil {
			// Send password reset email (silently fail to prevent enumeration)
			_ = s.emailService.SendPasswordResetEmail(u.Email, u.DisplayName, token)
		}
	}

	return nil
}

// ResetPassword resets a user's password using a valid reset token
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	if s.tokenService == nil {
		return errors.New("token service not available")
	}

	// Validate password
	if !ValidatePassword(newPassword) {
		return ErrWeakPassword
	}

	// Validate token and get user ID
	userID, err := s.tokenService.ValidatePasswordResetToken(ctx, token)
	if err != nil {
		return err
	}

	// Hash the new password
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// Update password in database
	err = s.userRepo.UpdatePassword(ctx, userID, hashedPassword)
	if err != nil {
		return errors.New("failed to update password")
	}

	// Delete all password reset tokens for this user
	_ = s.tokenService.DeletePasswordResetTokens(ctx, userID)

	return nil
}

// VerifyEmail verifies a user's email address using the provided token
func (s *Service) VerifyEmail(ctx context.Context, token string) (string, error) {
	if s.tokenService == nil {
		return "", errors.New("token service not available")
	}

	// Validate token and get user ID
	userID, err := s.tokenService.ValidateVerificationToken(ctx, token)
	if err != nil {
		return "", err
	}

	// Mark email as verified
	err = s.tokenService.MarkEmailAsVerified(ctx, userID)
	if err != nil {
		return "", err
	}

	return userID, nil
}

// ResendVerificationEmail resends the verification email to a user
func (s *Service) ResendVerificationEmail(ctx context.Context, email string) error {
	if s.tokenService == nil || s.emailService == nil {
		return errors.New("services not available")
	}

	// Get user by email (silently fail if not found to prevent enumeration)
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil
	}

	// Check if already verified
	if u.EmailVerifiedAt != nil {
		return nil
	}

	// Delete old verification tokens
	_ = s.tokenService.DeleteTokenByUser(ctx, u.ID, "email_verification")

	// Generate new token
	token, err := s.tokenService.CreateVerificationToken(ctx, u.ID)
	if err != nil {
		return nil
	}

	// Send verification email
	return s.emailService.SendVerificationEmail(u.Email, u.DisplayName, token)
}
