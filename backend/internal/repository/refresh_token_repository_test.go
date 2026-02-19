package repository_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/virallens/backend/internal/domain"
	"github.com/yourusername/virallens/backend/internal/repository"
)

func TestRefreshTokenRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewRefreshTokenRepository(db)

	// Create test user
	userID := uuid.New()
	createTestUser(t, db, userID, "testuser", "test@example.com")

	token := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     "test-refresh-token-12345",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	err := repo.Create(token)
	require.NoError(t, err)

	// Verify token was created
	found, err := repo.GetByToken(token.Token)
	require.NoError(t, err)
	assert.Equal(t, token.ID, found.ID)
	assert.Equal(t, token.UserID, found.UserID)
	assert.Equal(t, token.Token, found.Token)
}

func TestRefreshTokenRepository_GetByToken(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewRefreshTokenRepository(db)

	// Create test user
	userID := uuid.New()
	createTestUser(t, db, userID, "testuser", "test@example.com")

	token := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     "test-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	err := repo.Create(token)
	require.NoError(t, err)

	// Get by token
	found, err := repo.GetByToken("test-token")
	require.NoError(t, err)
	assert.Equal(t, token.ID, found.ID)
	assert.Equal(t, token.UserID, found.UserID)
}

func TestRefreshTokenRepository_GetByToken_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewRefreshTokenRepository(db)

	_, err := repo.GetByToken("nonexistent-token")
	assert.ErrorIs(t, err, domain.ErrInvalidToken)
}

func TestRefreshTokenRepository_DeleteByUserID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewRefreshTokenRepository(db)

	// Create test user
	userID := uuid.New()
	createTestUser(t, db, userID, "testuser", "test@example.com")

	// Create multiple tokens for the same user
	token1 := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     "token1",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
	token2 := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     "token2",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	require.NoError(t, repo.Create(token1))
	require.NoError(t, repo.Create(token2))

	// Delete all tokens for user
	err := repo.DeleteByUserID(userID)
	require.NoError(t, err)

	// Verify tokens are deleted
	_, err = repo.GetByToken("token1")
	assert.ErrorIs(t, err, domain.ErrInvalidToken)
	_, err = repo.GetByToken("token2")
	assert.ErrorIs(t, err, domain.ErrInvalidToken)
}

func TestRefreshTokenRepository_DeleteExpired(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewRefreshTokenRepository(db)

	// Create test user
	userID := uuid.New()
	createTestUser(t, db, userID, "testuser", "test@example.com")

	// Create expired token
	expiredToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Already expired
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	// Create valid token
	validToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     "valid-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	require.NoError(t, repo.Create(expiredToken))
	require.NoError(t, repo.Create(validToken))

	// Delete expired tokens
	err := repo.DeleteExpired()
	require.NoError(t, err)

	// Verify expired token is deleted
	_, err = repo.GetByToken("expired-token")
	assert.ErrorIs(t, err, domain.ErrInvalidToken)

	// Verify valid token still exists
	found, err := repo.GetByToken("valid-token")
	require.NoError(t, err)
	assert.Equal(t, validToken.Token, found.Token)
}
