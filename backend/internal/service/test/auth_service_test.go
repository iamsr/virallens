package service_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/virallens/backend/internal/domain"
	"github.com/yourusername/virallens/backend/internal/service"
	"github.com/yourusername/virallens/backend/test/mocks"
	"golang.org/x/crypto/bcrypt"
)


func TestAuthService_Register_Success(t *testing.T) {
	userRepo := mocks.NewMockUserRepository(t)
	tokenRepo := mocks.NewMockRefreshTokenRepository(t)
	jwtService := new(MockJWTService)

	authService := service.NewAuthService(userRepo, tokenRepo, jwtService)

	req := &service.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	// Mock: user doesn't exist
	userRepo.On("GetByUsername", req.Username).Return(nil, domain.ErrUserNotFound)
	userRepo.On("GetByEmail", req.Email).Return(nil, domain.ErrUserNotFound)

	// Mock: user creation succeeds
	userRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)

	// Mock: token generation
	jwtService.On("GenerateAccessToken", mock.AnythingOfType("uuid.UUID")).Return("access-token", nil)
	jwtService.On("GenerateRefreshToken").Return("refresh-token", nil)

	// Mock: refresh token storage
	tokenRepo.On("Create", mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

	resp, err := authService.Register(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, req.Username, resp.User.Username)
	assert.Equal(t, req.Email, resp.User.Email)
	assert.Equal(t, "access-token", resp.AccessToken)
	assert.Equal(t, "refresh-token", resp.RefreshToken)

	// Verify password was hashed
	err = bcrypt.CompareHashAndPassword([]byte(resp.User.PasswordHash), []byte(req.Password))
	assert.NoError(t, err)

	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	jwtService.AssertExpectations(t)
}

func TestAuthService_Register_UserAlreadyExists(t *testing.T) {
	userRepo := mocks.NewMockUserRepository(t)
	tokenRepo := mocks.NewMockRefreshTokenRepository(t)
	jwtService := new(MockJWTService)

	authService := service.NewAuthService(userRepo, tokenRepo, jwtService)

	req := &service.RegisterRequest{
		Username: "existinguser",
		Email:    "existing@example.com",
		Password: "password123",
	}

	existingUser := &domain.User{
		ID:       uuid.New(),
		Username: "existinguser",
		Email:    "existing@example.com",
	}

	// Mock: user already exists
	userRepo.On("GetByUsername", req.Username).Return(existingUser, nil)

	_, err := authService.Register(req)
	assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)

	userRepo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	userRepo := mocks.NewMockUserRepository(t)
	tokenRepo := mocks.NewMockRefreshTokenRepository(t)
	jwtService := new(MockJWTService)

	authService := service.NewAuthService(userRepo, tokenRepo, jwtService)

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
	}

	req := &service.LoginRequest{
		Username: "testuser",
		Password: password,
	}

	// Mock: user exists
	userRepo.On("GetByUsername", req.Username).Return(user, nil)

	// Mock: delete old tokens
	tokenRepo.On("DeleteByUserID", user.ID).Return(nil)

	// Mock: token generation
	jwtService.On("GenerateAccessToken", user.ID).Return("access-token", nil)
	jwtService.On("GenerateRefreshToken").Return("refresh-token", nil)

	// Mock: refresh token storage
	tokenRepo.On("Create", mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

	resp, err := authService.Login(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, user.Username, resp.User.Username)
	assert.Equal(t, "access-token", resp.AccessToken)
	assert.Equal(t, "refresh-token", resp.RefreshToken)

	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	jwtService.AssertExpectations(t)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	userRepo := mocks.NewMockUserRepository(t)
	tokenRepo := mocks.NewMockRefreshTokenRepository(t)
	jwtService := new(MockJWTService)

	authService := service.NewAuthService(userRepo, tokenRepo, jwtService)

	req := &service.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	// Mock: user not found
	userRepo.On("GetByUsername", req.Username).Return(nil, domain.ErrUserNotFound)

	_, err := authService.Login(req)
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)

	userRepo.AssertExpectations(t)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	userRepo := mocks.NewMockUserRepository(t)
	tokenRepo := mocks.NewMockRefreshTokenRepository(t)
	jwtService := new(MockJWTService)

	authService := service.NewAuthService(userRepo, tokenRepo, jwtService)

	correctPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		PasswordHash: string(hashedPassword),
	}

	req := &service.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	// Mock: user exists
	userRepo.On("GetByUsername", req.Username).Return(user, nil)

	_, err := authService.Login(req)
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)

	userRepo.AssertExpectations(t)
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	userRepo := mocks.NewMockUserRepository(t)
	tokenRepo := mocks.NewMockRefreshTokenRepository(t)
	jwtService := new(MockJWTService)

	authService := service.NewAuthService(userRepo, tokenRepo, jwtService)

	userID := uuid.New()
	user := &domain.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
	}

	refreshToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     "refresh-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	// Mock: refresh token exists and is valid
	tokenRepo.On("GetByToken", "refresh-token").Return(refreshToken, nil)

	// Mock: user exists
	userRepo.On("GetByID", userID).Return(user, nil)

	// Mock: delete old token
	tokenRepo.On("DeleteByUserID", userID).Return(nil)

	// Mock: new token generation
	jwtService.On("GenerateAccessToken", userID).Return("new-access-token", nil)
	jwtService.On("GenerateRefreshToken").Return("new-refresh-token", nil)

	// Mock: new refresh token storage
	tokenRepo.On("Create", mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

	resp, err := authService.RefreshToken("refresh-token")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, user.Username, resp.User.Username)
	assert.Equal(t, "new-access-token", resp.AccessToken)
	assert.Equal(t, "new-refresh-token", resp.RefreshToken)

	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	jwtService.AssertExpectations(t)
}

func TestAuthService_RefreshToken_Expired(t *testing.T) {
	userRepo := mocks.NewMockUserRepository(t)
	tokenRepo := mocks.NewMockRefreshTokenRepository(t)
	jwtService := new(MockJWTService)

	authService := service.NewAuthService(userRepo, tokenRepo, jwtService)

	userID := uuid.New()
	refreshToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	// Mock: refresh token exists but is expired
	tokenRepo.On("GetByToken", "expired-token").Return(refreshToken, nil)
	tokenRepo.On("DeleteByUserID", userID).Return(nil)

	_, err := authService.RefreshToken("expired-token")
	assert.ErrorIs(t, err, domain.ErrTokenExpired)

	tokenRepo.AssertExpectations(t)
}

func TestAuthService_Logout(t *testing.T) {
	userRepo := mocks.NewMockUserRepository(t)
	tokenRepo := mocks.NewMockRefreshTokenRepository(t)
	jwtService := new(MockJWTService)

	authService := service.NewAuthService(userRepo, tokenRepo, jwtService)

	userID := uuid.New()

	// Mock: delete tokens
	tokenRepo.On("DeleteByUserID", userID).Return(nil)

	err := authService.Logout(userID)
	require.NoError(t, err)

	tokenRepo.AssertExpectations(t)
}
