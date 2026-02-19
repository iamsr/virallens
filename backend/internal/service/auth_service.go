package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/virallens/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Username string
	Email    string
	Password string
}

type LoginRequest struct {
	Username string
	Password string
}

type AuthResponse struct {
	User         *domain.User
	AccessToken  string
	RefreshToken string
}

type AuthService interface {
	Register(req *RegisterRequest) (*AuthResponse, error)
	Login(req *LoginRequest) (*AuthResponse, error)
	RefreshToken(refreshToken string) (*AuthResponse, error)
	Logout(userID uuid.UUID) error
}

type authService struct {
	userRepo         domain.UserRepository
	refreshTokenRepo domain.RefreshTokenRepository
	jwtService       JWTService
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo domain.UserRepository,
	refreshTokenRepo domain.RefreshTokenRepository,
	jwtService JWTService,
) AuthService {
	return &authService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
	}
}

func (s *authService) Register(req *RegisterRequest) (*AuthResponse, error) {
	// Check if user already exists
	existingUser, _ := s.userRepo.GetByUsername(req.Username)
	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	existingUser, _ = s.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &domain.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Generate tokens
	return s.generateAuthResponse(user)
}

func (s *authService) Login(req *LoginRequest) (*AuthResponse, error) {
	// Get user by username
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Delete old refresh tokens
	_ = s.refreshTokenRepo.DeleteByUserID(user.ID)

	// Generate new tokens
	return s.generateAuthResponse(user)
}

func (s *authService) RefreshToken(refreshToken string) (*AuthResponse, error) {
	// Get refresh token from database
	token, err := s.refreshTokenRepo.GetByToken(refreshToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	// Check if token is expired
	if token.ExpiresAt.Before(time.Now()) {
		_ = s.refreshTokenRepo.DeleteByUserID(token.UserID)
		return nil, domain.ErrTokenExpired
	}

	// Get user
	user, err := s.userRepo.GetByID(token.UserID)
	if err != nil {
		return nil, err
	}

	// Delete old refresh token
	_ = s.refreshTokenRepo.DeleteByUserID(user.ID)

	// Generate new tokens
	return s.generateAuthResponse(user)
}

func (s *authService) Logout(userID uuid.UUID) error {
	// Delete all refresh tokens for the user
	return s.refreshTokenRepo.DeleteByUserID(userID)
}

func (s *authService) generateAuthResponse(user *domain.User) (*AuthResponse, error) {
	// Generate access token
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := s.jwtService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Store refresh token in database
	token := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
		CreatedAt: time.Now(),
	}

	if err := s.refreshTokenRepo.Create(token); err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
