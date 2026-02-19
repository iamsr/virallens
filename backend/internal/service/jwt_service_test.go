package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTService_GenerateAccessToken(t *testing.T) {
	jwtService := NewJWTService("test-secret-key", 15*time.Minute, 24*time.Hour)

	userID := uuid.New()
	token, err := jwtService.GenerateAccessToken(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token can be validated
	claims, err := jwtService.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
}

func TestJWTService_GenerateRefreshToken(t *testing.T) {
	jwtService := NewJWTService("test-secret-key", 15*time.Minute, 24*time.Hour)

	token1, err := jwtService.GenerateRefreshToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token1)

	token2, err := jwtService.GenerateRefreshToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token2)

	// Tokens should be unique
	assert.NotEqual(t, token1, token2)
}

func TestJWTService_ValidateAccessToken_Valid(t *testing.T) {
	jwtService := NewJWTService("test-secret-key", 15*time.Minute, 24*time.Hour)

	userID := uuid.New()
	token, err := jwtService.GenerateAccessToken(userID)
	require.NoError(t, err)

	claims, err := jwtService.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
}

func TestJWTService_ValidateAccessToken_Invalid(t *testing.T) {
	jwtService := NewJWTService("test-secret-key", 15*time.Minute, 24*time.Hour)

	// Test with invalid token
	_, err := jwtService.ValidateAccessToken("invalid-token")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestJWTService_ValidateAccessToken_WrongSecret(t *testing.T) {
	jwtService1 := NewJWTService("secret-key-1", 15*time.Minute, 24*time.Hour)
	jwtService2 := NewJWTService("secret-key-2", 15*time.Minute, 24*time.Hour)

	userID := uuid.New()
	token, err := jwtService1.GenerateAccessToken(userID)
	require.NoError(t, err)

	// Try to validate with different secret
	_, err = jwtService2.ValidateAccessToken(token)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestJWTService_ValidateAccessToken_Expired(t *testing.T) {
	// Create service with very short expiration
	jwtService := NewJWTService("test-secret-key", 1*time.Millisecond, 24*time.Hour)

	userID := uuid.New()
	token, err := jwtService.GenerateAccessToken(userID)
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	_, err = jwtService.ValidateAccessToken(token)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestJWTService_TokenContainsClaims(t *testing.T) {
	jwtService := NewJWTService("test-secret-key", 15*time.Minute, 24*time.Hour)

	userID := uuid.New()
	token, err := jwtService.GenerateAccessToken(userID)
	require.NoError(t, err)

	claims, err := jwtService.ValidateAccessToken(token)
	require.NoError(t, err)

	// Verify all claims are present
	assert.Equal(t, userID, claims.UserID)
	assert.True(t, claims.ExpiresAt.After(time.Now()))
	assert.True(t, claims.IssuedAt.Before(time.Now().Add(1*time.Second)))
	assert.True(t, claims.NotBefore.Before(time.Now().Add(1*time.Second)))
}
