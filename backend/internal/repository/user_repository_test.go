package repository_test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/virallens/backend/internal/domain"
	"github.com/yourusername/virallens/backend/internal/repository"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/virallens_test?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Skipf("Failed to connect to test database: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		t.Skipf("Test database not available: %v", err)
	}

	// Clean up tables before each test
	_, err = db.Exec(`
		TRUNCATE users, conversations, conversation_participants, 
		groups, group_members, messages, refresh_tokens CASCADE
	`)
	if err != nil {
		t.Fatalf("Failed to clean up tables: %v", err)
	}

	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewUserRepository(db)

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(user)
	require.NoError(t, err)

	// Verify user was created
	found, err := repo.GetByID(user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Username, found.Username)
	assert.Equal(t, user.Email, found.Email)
	assert.Equal(t, user.PasswordHash, found.PasswordHash)
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewUserRepository(db)

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(user)
	require.NoError(t, err)

	found, err := repo.GetByUsername("testuser")
	require.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, user.Email, found.Email)
}

func TestUserRepository_GetByUsername_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewUserRepository(db)

	_, err := repo.GetByUsername("nonexistent")
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewUserRepository(db)

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(user)
	require.NoError(t, err)

	found, err := repo.GetByEmail("test@example.com")
	require.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, user.Username, found.Username)
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewUserRepository(db)

	// Create multiple users
	user1 := &domain.User{
		ID:           uuid.New(),
		Username:     "user1",
		Email:        "user1@example.com",
		PasswordHash: "hash1",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	user2 := &domain.User{
		ID:           uuid.New(),
		Username:     "user2",
		Email:        "user2@example.com",
		PasswordHash: "hash2",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	require.NoError(t, repo.Create(user1))
	require.NoError(t, repo.Create(user2))

	users, err := repo.List()
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestUserRepository_Create_DuplicateUsername(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewUserRepository(db)

	user1 := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test1@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	require.NoError(t, repo.Create(user1))

	// Try to create another user with same username
	user2 := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test2@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(user2)
	assert.Error(t, err)
}
