package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/virallens/backend/models"
	"github.com/yourusername/virallens/backend/modules/auth/dto"
	"github.com/yourusername/virallens/backend/modules/user"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrTokenExpired       = errors.New("refresh token expired")
	ErrInvalidToken       = errors.New("invalid refresh token")
)

type AuthResponse struct {
	User         *models.User
	AccessToken  string
	RefreshToken string
}

type Service interface {
	Register(req *dto.RegisterRequest) (*AuthResponse, error)
	Login(req *dto.LoginRequest) (*AuthResponse, error)
	RefreshToken(refreshToken string) (*AuthResponse, error)
	Logout(userID uuid.UUID) error
}

type service struct {
	userRepo         user.Repository
	refreshTokenRepo RefreshTokenRepository
	jwtService       JWTService
}

func NewService(
	userRepo user.Repository,
	refreshTokenRepo RefreshTokenRepository,
	jwtService JWTService,
) Service {
	return &service{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
	}
}

func (s *service) Register(req *dto.RegisterRequest) (*AuthResponse, error) {
	existingUser, _ := s.userRepo.GetByUsername(req.Username)
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}
	existingUser, _ = s.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &models.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := s.userRepo.Create(u); err != nil {
		return nil, err
	}

	return s.generateAuthResponse(u)
}

func (s *service) Login(req *dto.LoginRequest) (*AuthResponse, error) {
	u, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	_ = s.refreshTokenRepo.DeleteByUserID(u.ID)
	return s.generateAuthResponse(u)
}

func (s *service) RefreshToken(refreshToken string) (*AuthResponse, error) {
	token, err := s.refreshTokenRepo.GetByToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if token.ExpiresAt.Before(time.Now()) {
		_ = s.refreshTokenRepo.DeleteByUserID(token.UserID)
		return nil, ErrTokenExpired
	}

	u, err := s.userRepo.GetByID(token.UserID)
	if err != nil {
		return nil, err
	}

	_ = s.refreshTokenRepo.DeleteByUserID(u.ID)
	return s.generateAuthResponse(u)
}

func (s *service) Logout(userID uuid.UUID) error {
	return s.refreshTokenRepo.DeleteByUserID(userID)
}

func (s *service) generateAuthResponse(u *models.User) (*AuthResponse, error) {
	accessToken, err := s.jwtService.GenerateAccessToken(u.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	token := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    u.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.refreshTokenRepo.Create(token); err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:         u,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
