package repository_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/virallens/backend/internal/domain"
	"github.com/yourusername/virallens/backend/internal/repository"
	"github.com/yourusername/virallens/backend/test/testutil"
)

// ============================================================================

func TestConversationRepository_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	repo := repository.NewConversationRepository(db)

	// Create test users first
	user1ID := uuid.New()
	user2ID := uuid.New()
	createTestUser(t, db, user1ID, "user1", "user1@example.com")
	createTestUser(t, db, user2ID, "user2", "user2@example.com")

	conv := &domain.Conversation{
		ID:           uuid.New(),
		Participants: []uuid.UUID{user1ID, user2ID},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(conv)
	require.NoError(t, err)

	// Verify conversation was created
	found, err := repo.GetByID(conv.ID)
	require.NoError(t, err)
	assert.Equal(t, conv.ID, found.ID)
	assert.ElementsMatch(t, conv.Participants, found.Participants)
}

func TestConversationRepository_GetByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	repo := repository.NewConversationRepository(db)

	// Create test users
	user1ID := uuid.New()
	user2ID := uuid.New()
	createTestUser(t, db, user1ID, "user1", "user1@example.com")
	createTestUser(t, db, user2ID, "user2", "user2@example.com")

	conv := &domain.Conversation{
		ID:           uuid.New(),
		Participants: []uuid.UUID{user1ID, user2ID},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(conv)
	require.NoError(t, err)

	found, err := repo.GetByID(conv.ID)
	require.NoError(t, err)
	assert.Equal(t, conv.ID, found.ID)
	assert.Len(t, found.Participants, 2)
}

func TestConversationRepository_GetByID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	repo := repository.NewConversationRepository(db)

	_, err := repo.GetByID(uuid.New())
	assert.ErrorIs(t, err, domain.ErrConversationNotFound)
}

func TestConversationRepository_GetByParticipants(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	repo := repository.NewConversationRepository(db)

	// Create test users
	user1ID := uuid.New()
	user2ID := uuid.New()
	createTestUser(t, db, user1ID, "user1", "user1@example.com")
	createTestUser(t, db, user2ID, "user2", "user2@example.com")

	conv := &domain.Conversation{
		ID:           uuid.New(),
		Participants: []uuid.UUID{user1ID, user2ID},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(conv)
	require.NoError(t, err)

	// Should find conversation regardless of participant order
	found, err := repo.GetByParticipants(user1ID, user2ID)
	require.NoError(t, err)
	assert.Equal(t, conv.ID, found.ID)

	found, err = repo.GetByParticipants(user2ID, user1ID)
	require.NoError(t, err)
	assert.Equal(t, conv.ID, found.ID)
}

func TestConversationRepository_GetByParticipants_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	repo := repository.NewConversationRepository(db)

	user1ID := uuid.New()
	user2ID := uuid.New()

	_, err := repo.GetByParticipants(user1ID, user2ID)
	assert.ErrorIs(t, err, domain.ErrConversationNotFound)
}

func TestConversationRepository_ListByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	repo := repository.NewConversationRepository(db)

	// Create test users
	user1ID := uuid.New()
	user2ID := uuid.New()
	user3ID := uuid.New()
	createTestUser(t, db, user1ID, "user1", "user1@example.com")
	createTestUser(t, db, user2ID, "user2", "user2@example.com")
	createTestUser(t, db, user3ID, "user3", "user3@example.com")

	// Create conversations
	conv1 := &domain.Conversation{
		ID:           uuid.New(),
		Participants: []uuid.UUID{user1ID, user2ID},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	conv2 := &domain.Conversation{
		ID:           uuid.New(),
		Participants: []uuid.UUID{user1ID, user3ID},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	require.NoError(t, repo.Create(conv1))
	require.NoError(t, repo.Create(conv2))

	// User1 should have 2 conversations
	convs, err := repo.ListByUserID(user1ID)
	require.NoError(t, err)
	assert.Len(t, convs, 2)

	// User2 should have 1 conversation
	convs, err = repo.ListByUserID(user2ID)
	require.NoError(t, err)
	assert.Len(t, convs, 1)
	assert.Equal(t, conv1.ID, convs[0].ID)
}

func TestConversationRepository_AddParticipant(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	repo := repository.NewConversationRepository(db)

	// Create test users
	user1ID := uuid.New()
	user2ID := uuid.New()
	user3ID := uuid.New()
	createTestUser(t, db, user1ID, "user1", "user1@example.com")
	createTestUser(t, db, user2ID, "user2", "user2@example.com")
	createTestUser(t, db, user3ID, "user3", "user3@example.com")

	conv := &domain.Conversation{
		ID:           uuid.New(),
		Participants: []uuid.UUID{user1ID, user2ID},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(conv)
	require.NoError(t, err)

	// Add third participant
	err = repo.AddParticipant(conv.ID, user3ID)
	require.NoError(t, err)

	// Verify participant was added
	found, err := repo.GetByID(conv.ID)
	require.NoError(t, err)
	assert.Len(t, found.Participants, 3)
	assert.Contains(t, found.Participants, user3ID)
}
