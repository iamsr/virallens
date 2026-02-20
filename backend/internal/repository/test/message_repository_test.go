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

func TestMessageRepository_Create_Conversation(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	repo := repository.NewMessageRepository(db)

	// Create test users
	user1ID := uuid.New()
	user2ID := uuid.New()
	createTestUser(t, db, user1ID, "user1", "user1@example.com")
	createTestUser(t, db, user2ID, "user2", "user2@example.com")

	// Create conversation
	convID := createTestConversation(t, db, user1ID, user2ID)

	message := &domain.Message{
		ID:             uuid.New(),
		SenderID:       user1ID,
		ConversationID: &convID,
		Content:        "Hello, this is a test message",
		Type:           domain.MessageTypeConversation,
		CreatedAt:      time.Now(),
	}

	err := repo.Create(message)
	require.NoError(t, err)

	// Verify message was created
	found, err := repo.GetByID(message.ID)
	require.NoError(t, err)
	assert.Equal(t, message.ID, found.ID)
	assert.Equal(t, message.SenderID, found.SenderID)
	assert.Equal(t, message.Content, found.Content)
	assert.Equal(t, message.Type, found.Type)
	assert.NotNil(t, found.ConversationID)
	assert.Equal(t, *message.ConversationID, *found.ConversationID)
	assert.Nil(t, found.GroupID)
}

func TestMessageRepository_Create_Group(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	repo := repository.NewMessageRepository(db)

	// Create test users
	user1ID := uuid.New()
	user2ID := uuid.New()
	createTestUser(t, db, user1ID, "user1", "user1@example.com")
	createTestUser(t, db, user2ID, "user2", "user2@example.com")

	// Create group
	groupID := createTestGroup(t, db, "Test Group", user1ID, []uuid.UUID{user1ID, user2ID})

	message := &domain.Message{
		ID:        uuid.New(),
		SenderID:  user1ID,
		GroupID:   &groupID,
		Content:   "Hello group!",
		Type:      domain.MessageTypeGroup,
		CreatedAt: time.Now(),
	}

	err := repo.Create(message)
	require.NoError(t, err)

	// Verify message was created
	found, err := repo.GetByID(message.ID)
	require.NoError(t, err)
	assert.Equal(t, message.ID, found.ID)
	assert.Equal(t, message.SenderID, found.SenderID)
	assert.Equal(t, message.Content, found.Content)
	assert.Equal(t, message.Type, found.Type)
	assert.NotNil(t, found.GroupID)
	assert.Equal(t, *message.GroupID, *found.GroupID)
	assert.Nil(t, found.ConversationID)
}

func TestMessageRepository_GetByID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	repo := repository.NewMessageRepository(db)

	_, err := repo.GetByID(uuid.New())
	assert.ErrorIs(t, err, domain.ErrMessageNotFound)
}

func TestMessageRepository_ListByConversationID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	repo := repository.NewMessageRepository(db)

	// Create test users
	user1ID := uuid.New()
	user2ID := uuid.New()
	createTestUser(t, db, user1ID, "user1", "user1@example.com")
	createTestUser(t, db, user2ID, "user2", "user2@example.com")

	// Create conversation
	convID := createTestConversation(t, db, user1ID, user2ID)

	// Create messages with known timestamps
	now := time.Now()
	msg1 := &domain.Message{
		ID:             uuid.New(),
		SenderID:       user1ID,
		ConversationID: &convID,
		Content:        "Message 1",
		Type:           domain.MessageTypeConversation,
		CreatedAt:      now.Add(-2 * time.Minute),
	}
	msg2 := &domain.Message{
		ID:             uuid.New(),
		SenderID:       user2ID,
		ConversationID: &convID,
		Content:        "Message 2",
		Type:           domain.MessageTypeConversation,
		CreatedAt:      now.Add(-1 * time.Minute),
	}
	msg3 := &domain.Message{
		ID:             uuid.New(),
		SenderID:       user1ID,
		ConversationID: &convID,
		Content:        "Message 3",
		Type:           domain.MessageTypeConversation,
		CreatedAt:      now,
	}

	require.NoError(t, repo.Create(msg1))
	require.NoError(t, repo.Create(msg2))
	require.NoError(t, repo.Create(msg3))

	// Get all messages (no cursor)
	messages, err := repo.ListByConversationID(convID, nil, 10)
	require.NoError(t, err)
	assert.Len(t, messages, 3)
	// Should be in descending order (newest first)
	assert.Equal(t, msg3.ID, messages[0].ID)
	assert.Equal(t, msg2.ID, messages[1].ID)
	assert.Equal(t, msg1.ID, messages[2].ID)

	// Get messages with cursor (before msg3)
	cursor := msg3.CreatedAt
	messages, err = repo.ListByConversationID(convID, &cursor, 10)
	require.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, msg2.ID, messages[0].ID)
	assert.Equal(t, msg1.ID, messages[1].ID)

	// Get messages with limit
	messages, err = repo.ListByConversationID(convID, nil, 2)
	require.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, msg3.ID, messages[0].ID)
	assert.Equal(t, msg2.ID, messages[1].ID)
}

func TestMessageRepository_ListByGroupID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	repo := repository.NewMessageRepository(db)

	// Create test users
	user1ID := uuid.New()
	user2ID := uuid.New()
	createTestUser(t, db, user1ID, "user1", "user1@example.com")
	createTestUser(t, db, user2ID, "user2", "user2@example.com")

	// Create group
	groupID := createTestGroup(t, db, "Test Group", user1ID, []uuid.UUID{user1ID, user2ID})

	// Create messages
	now := time.Now()
	msg1 := &domain.Message{
		ID:        uuid.New(),
		SenderID:  user1ID,
		GroupID:   &groupID,
		Content:   "Group message 1",
		Type:      domain.MessageTypeGroup,
		CreatedAt: now.Add(-1 * time.Minute),
	}
	msg2 := &domain.Message{
		ID:        uuid.New(),
		SenderID:  user2ID,
		GroupID:   &groupID,
		Content:   "Group message 2",
		Type:      domain.MessageTypeGroup,
		CreatedAt: now,
	}

	require.NoError(t, repo.Create(msg1))
	require.NoError(t, repo.Create(msg2))

	// Get all messages
	messages, err := repo.ListByGroupID(groupID, nil, 10)
	require.NoError(t, err)
	assert.Len(t, messages, 2)
	// Should be in descending order (newest first)
	assert.Equal(t, msg2.ID, messages[0].ID)
	assert.Equal(t, msg1.ID, messages[1].ID)

	// Get messages with cursor
	cursor := msg2.CreatedAt
	messages, err = repo.ListByGroupID(groupID, &cursor, 10)
	require.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, msg1.ID, messages[0].ID)
}
