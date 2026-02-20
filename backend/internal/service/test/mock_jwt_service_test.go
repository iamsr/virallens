package service_test

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/yourusername/virallens/backend/internal/service"
)

// MockJWTService - JWT service is not a domain interface, so we mock it manually
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateAccessToken(userID uuid.UUID) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) GenerateRefreshToken() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) ValidateAccessToken(tokenString string) (*service.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.Claims), args.Error(1)
}
