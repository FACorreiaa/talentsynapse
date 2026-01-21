package auth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/FACorreiaa/skillsphere-pwa/internal/app/auth"
	"github.com/FACorreiaa/skillsphere-pwa/internal/app/user"
)

// MockUserRepository is a mock implementation of user.RepositoryInterface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, email, username, hashedPassword, displayName string) (*user.User, error) {
	args := m.Called(ctx, email, username, hashedPassword, displayName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestAuthService_Login_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	svc := auth.NewService(mockRepo)
	ctx := context.Background()

	// Create a test user with hashed password
	hashedPwd, _ := auth.HashPassword("correctpassword")
	testUser := &user.User{
		ID:             "user-123",
		Email:          "test@example.com",
		Username:       "testuser",
		HashedPassword: hashedPwd,
		DisplayName:    "Test User",
	}

	mockRepo.On("GetByEmail", ctx, "test@example.com").Return(testUser, nil)
	mockRepo.On("UpdateLastLogin", ctx, "user-123").Return(nil)

	// Act
	input := user.LoginInput{
		Email:    "test@example.com",
		Password: "correctpassword",
	}
	result, err := svc.Login(ctx, input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "user-123", result.ID)
	assert.Equal(t, "test@example.com", result.Email)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	svc := auth.NewService(mockRepo)
	ctx := context.Background()

	hashedPwd, _ := auth.HashPassword("correctpassword")
	testUser := &user.User{
		ID:             "user-123",
		Email:          "test@example.com",
		HashedPassword: hashedPwd,
	}

	mockRepo.On("GetByEmail", ctx, "test@example.com").Return(testUser, nil)

	// Act
	input := user.LoginInput{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	result, err := svc.Login(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, auth.ErrInvalidCredentials, err)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	svc := auth.NewService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByEmail", ctx, "nonexistent@example.com").Return(nil, user.ErrUserNotFound)

	// Act
	input := user.LoginInput{
		Email:    "nonexistent@example.com",
		Password: "anypassword",
	}
	result, err := svc.Login(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, auth.ErrInvalidCredentials, err)
}

func TestAuthService_Register_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	svc := auth.NewService(mockRepo)
	ctx := context.Background()

	mockRepo.On("Create", ctx, "newuser@example.com", "newuser", mock.Anything, "New User").
		Return(&user.User{
			ID:          "user-456",
			Email:       "newuser@example.com",
			Username:    "newuser",
			DisplayName: "New User",
		}, nil)

	// Act
	input := user.RegisterInput{
		Email:           "newuser@example.com",
		Username:        "newuser",
		Password:        "securepassword123",
		ConfirmPassword: "securepassword123",
		DisplayName:     "New User",
	}
	result, err := svc.Register(ctx, input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "user-456", result.ID)
	assert.Equal(t, "newuser@example.com", result.Email)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Register_PasswordMismatch(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	svc := auth.NewService(mockRepo)
	ctx := context.Background()

	// Act
	input := user.RegisterInput{
		Email:           "newuser@example.com",
		Username:        "newuser",
		Password:        "password123",
		ConfirmPassword: "differentpassword",
		DisplayName:     "New User",
	}
	result, err := svc.Register(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, auth.ErrPasswordMismatch, err)
}

func TestAuthService_Register_WeakPassword(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	svc := auth.NewService(mockRepo)
	ctx := context.Background()

	// Act
	input := user.RegisterInput{
		Email:           "newuser@example.com",
		Username:        "newuser",
		Password:        "short",
		ConfirmPassword: "short",
		DisplayName:     "New User",
	}
	result, err := svc.Register(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, auth.ErrWeakPassword, err)
}

func TestAuthService_Register_InvalidEmail(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	svc := auth.NewService(mockRepo)
	ctx := context.Background()

	// Act
	input := user.RegisterInput{
		Email:           "invalidemail",
		Username:        "newuser",
		Password:        "securepassword123",
		ConfirmPassword: "securepassword123",
		DisplayName:     "New User",
	}
	result, err := svc.Register(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, auth.ErrInvalidEmail, err)
}

func TestAuthService_Register_EmailAlreadyExists(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	svc := auth.NewService(mockRepo)
	ctx := context.Background()

	mockRepo.On("Create", ctx, "existing@example.com", "newuser", mock.Anything, "New User").
		Return(nil, user.ErrEmailAlreadyExists)

	// Act
	input := user.RegisterInput{
		Email:           "existing@example.com",
		Username:        "newuser",
		Password:        "securepassword123",
		ConfirmPassword: "securepassword123",
		DisplayName:     "New User",
	}
	result, err := svc.Register(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, user.ErrEmailAlreadyExists, err)
}

func TestPasswordHashing(t *testing.T) {
	password := "mysecretpassword"

	// Test hashing
	hash, err := auth.HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	// Test checking correct password
	assert.True(t, auth.CheckPassword(hash, password))

	// Test checking wrong password
	assert.False(t, auth.CheckPassword(hash, "wrongpassword"))
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{"valid password", "password123", true},
		{"exactly 8 chars", "12345678", true},
		{"too short", "short", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auth.ValidatePassword(tt.password)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{"valid email", "test@example.com", true},
		{"valid with subdomain", "test@mail.example.com", true},
		{"missing @", "testexample.com", false},
		{"missing domain", "test@", false},
		{"missing local part", "@example.com", false},
		{"missing dot after @", "test@example", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auth.ValidateEmail(tt.email)
			assert.Equal(t, tt.valid, result)
		})
	}
}
