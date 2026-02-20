package integration

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/virallens/backend/internal/domain"
	"github.com/yourusername/virallens/backend/internal/repository"
	"github.com/yourusername/virallens/backend/test/testutil"
)

// ============================================================================
// User Repository Tests
// ============================================================================

func TestUserRepository_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

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
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

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
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	repo := repository.NewUserRepository(db)

	_, err := repo.GetByUsername("nonexistent")
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

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
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

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
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

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

// ============================================================================
// Conversation Repository Tests
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

// ============================================================================
// Group Repository Tests
// ============================================================================

func TestGroupRepository_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	repo := repository.NewGroupRepository(db)

	// Create test users
	creatorID := uuid.New()
	member1ID := uuid.New()
	createTestUser(t, db, creatorID, "creator", "creator@example.com")
	createTestUser(t, db, member1ID, "member1", "member1@example.com")

	group := &domain.Group{
		ID:        uuid.New(),
		Name:      "Test Group",
		CreatedBy: creatorID,
		Members:   []uuid.UUID{creatorID, member1ID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(group)
	require.NoError(t, err)

	// Verify group was created
	found, err := repo.GetByID(group.ID)
	require.NoError(t, err)
	assert.Equal(t, group.ID, found.ID)
	assert.Equal(t, group.Name, found.Name)
	assert.Equal(t, group.CreatedBy, found.CreatedBy)
	assert.ElementsMatch(t, group.Members, found.Members)
}

func TestGroupRepository_GetByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	repo := repository.NewGroupRepository(db)

	// Create test user
	creatorID := uuid.New()
	createTestUser(t, db, creatorID, "creator", "creator@example.com")

	group := &domain.Group{
		ID:        uuid.New(),
		Name:      "Test Group",
		CreatedBy: creatorID,
		Members:   []uuid.UUID{creatorID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(group)
	require.NoError(t, err)

	found, err := repo.GetByID(group.ID)
	require.NoError(t, err)
	assert.Equal(t, group.ID, found.ID)
	assert.Equal(t, group.Name, found.Name)
	assert.Equal(t, group.CreatedBy, found.CreatedBy)
}

func TestGroupRepository_GetByID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	repo := repository.NewGroupRepository(db)

	_, err := repo.GetByID(uuid.New())
	assert.ErrorIs(t, err, domain.ErrGroupNotFound)
}

func TestGroupRepository_ListByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	repo := repository.NewGroupRepository(db)

	// Create test users
	user1ID := uuid.New()
	user2ID := uuid.New()
	user3ID := uuid.New()
	createTestUser(t, db, user1ID, "user1", "user1@example.com")
	createTestUser(t, db, user2ID, "user2", "user2@example.com")
	createTestUser(t, db, user3ID, "user3", "user3@example.com")

	// Create groups
	group1 := &domain.Group{
		ID:        uuid.New(),
		Name:      "Group 1",
		CreatedBy: user1ID,
		Members:   []uuid.UUID{user1ID, user2ID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	group2 := &domain.Group{
		ID:        uuid.New(),
		Name:      "Group 2",
		CreatedBy: user2ID,
		Members:   []uuid.UUID{user1ID, user2ID, user3ID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	require.NoError(t, repo.Create(group1))
	require.NoError(t, repo.Create(group2))

	// User1 should be in 2 groups
	groups, err := repo.ListByUserID(user1ID)
	require.NoError(t, err)
	assert.Len(t, groups, 2)

	// User3 should be in 1 group
	groups, err = repo.ListByUserID(user3ID)
	require.NoError(t, err)
	assert.Len(t, groups, 1)
	assert.Equal(t, group2.ID, groups[0].ID)
}

func TestGroupRepository_AddMember(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	repo := repository.NewGroupRepository(db)

	// Create test users
	creatorID := uuid.New()
	newMemberID := uuid.New()
	createTestUser(t, db, creatorID, "creator", "creator@example.com")
	createTestUser(t, db, newMemberID, "newmember", "newmember@example.com")

	group := &domain.Group{
		ID:        uuid.New(),
		Name:      "Test Group",
		CreatedBy: creatorID,
		Members:   []uuid.UUID{creatorID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(group)
	require.NoError(t, err)

	// Add new member
	err = repo.AddMember(group.ID, newMemberID)
	require.NoError(t, err)

	// Verify member was added
	found, err := repo.GetByID(group.ID)
	require.NoError(t, err)
	assert.Len(t, found.Members, 2)
	assert.Contains(t, found.Members, newMemberID)
}

func TestGroupRepository_RemoveMember(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	repo := repository.NewGroupRepository(db)

	// Create test users
	creatorID := uuid.New()
	memberID := uuid.New()
	createTestUser(t, db, creatorID, "creator", "creator@example.com")
	createTestUser(t, db, memberID, "member", "member@example.com")

	group := &domain.Group{
		ID:        uuid.New(),
		Name:      "Test Group",
		CreatedBy: creatorID,
		Members:   []uuid.UUID{creatorID, memberID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(group)
	require.NoError(t, err)

	// Remove member
	err = repo.RemoveMember(group.ID, memberID)
	require.NoError(t, err)

	// Verify member was removed
	found, err := repo.GetByID(group.ID)
	require.NoError(t, err)
	assert.Len(t, found.Members, 1)
	assert.NotContains(t, found.Members, memberID)
}

func TestGroupRepository_IsMember(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	repo := repository.NewGroupRepository(db)

	// Create test users
	creatorID := uuid.New()
	memberID := uuid.New()
	nonMemberID := uuid.New()
	createTestUser(t, db, creatorID, "creator", "creator@example.com")
	createTestUser(t, db, memberID, "member", "member@example.com")
	createTestUser(t, db, nonMemberID, "nonmember", "nonmember@example.com")

	group := &domain.Group{
		ID:        uuid.New(),
		Name:      "Test Group",
		CreatedBy: creatorID,
		Members:   []uuid.UUID{creatorID, memberID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(group)
	require.NoError(t, err)

	// Check if creator is member
	isMember, err := repo.IsMember(group.ID, creatorID)
	require.NoError(t, err)
	assert.True(t, isMember)

	// Check if member is member
	isMember, err = repo.IsMember(group.ID, memberID)
	require.NoError(t, err)
	assert.True(t, isMember)

	// Check if non-member is member
	isMember, err = repo.IsMember(group.ID, nonMemberID)
	require.NoError(t, err)
	assert.False(t, isMember)
}

// ============================================================================
// Message Repository Tests
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

// ============================================================================
// Refresh Token Repository Tests
// ============================================================================

func TestRefreshTokenRepository_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

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
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

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
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	repo := repository.NewRefreshTokenRepository(db)

	_, err := repo.GetByToken("nonexistent-token")
	assert.ErrorIs(t, err, domain.ErrInvalidToken)
}

func TestRefreshTokenRepository_DeleteByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

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
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

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

// ============================================================================
// Integration Tests
// ============================================================================

// TestIntegration_ConversationWithMessages tests the integration between
// User, Conversation, and Message repositories
func TestIntegration_ConversationWithMessages(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	convRepo := repository.NewConversationRepository(db)
	msgRepo := repository.NewMessageRepository(db)

	// Create users
	user1 := &domain.User{
		ID:           uuid.New(),
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hash1",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	user2 := &domain.User{
		ID:           uuid.New(),
		Username:     "bob",
		Email:        "bob@example.com",
		PasswordHash: "hash2",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	require.NoError(t, userRepo.Create(user1))
	require.NoError(t, userRepo.Create(user2))

	// Create conversation between users
	conv := &domain.Conversation{
		ID:           uuid.New(),
		Participants: []uuid.UUID{user1.ID, user2.ID},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, convRepo.Create(conv))

	// Send messages in the conversation
	msg1 := &domain.Message{
		ID:             uuid.New(),
		SenderID:       user1.ID,
		ConversationID: &conv.ID,
		Content:        "Hello Bob!",
		Type:           domain.MessageTypeConversation,
		CreatedAt:      time.Now(),
	}
	msg2 := &domain.Message{
		ID:             uuid.New(),
		SenderID:       user2.ID,
		ConversationID: &conv.ID,
		Content:        "Hi Alice!",
		Type:           domain.MessageTypeConversation,
		CreatedAt:      time.Now().Add(1 * time.Second),
	}

	require.NoError(t, msgRepo.Create(msg1))
	require.NoError(t, msgRepo.Create(msg2))

	// List messages by conversation
	messages, err := msgRepo.ListByConversationID(conv.ID, nil, 10)
	require.NoError(t, err)
	assert.Len(t, messages, 2)

	// Verify message order (newest first)
	assert.Equal(t, msg2.ID, messages[0].ID)
	assert.Equal(t, msg1.ID, messages[1].ID)

	// Verify conversation can be retrieved
	foundConv, err := convRepo.GetByID(conv.ID)
	require.NoError(t, err)
	assert.Len(t, foundConv.Participants, 2)

	// Verify users can be retrieved
	foundUser1, err := userRepo.GetByID(user1.ID)
	require.NoError(t, err)
	assert.Equal(t, "alice", foundUser1.Username)
}

// TestIntegration_GroupWithMessages tests the integration between
// User, Group, and Message repositories
func TestIntegration_GroupWithMessages(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	msgRepo := repository.NewMessageRepository(db)

	// Create users
	creator := &domain.User{
		ID:           uuid.New(),
		Username:     "creator",
		Email:        "creator@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	member1 := &domain.User{
		ID:           uuid.New(),
		Username:     "member1",
		Email:        "member1@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	member2 := &domain.User{
		ID:           uuid.New(),
		Username:     "member2",
		Email:        "member2@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	require.NoError(t, userRepo.Create(creator))
	require.NoError(t, userRepo.Create(member1))
	require.NoError(t, userRepo.Create(member2))

	// Create group
	group := &domain.Group{
		ID:        uuid.New(),
		Name:      "Team Chat",
		CreatedBy: creator.ID,
		Members:   []uuid.UUID{creator.ID, member1.ID, member2.ID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, groupRepo.Create(group))

	// Send messages in the group
	msg1 := &domain.Message{
		ID:        uuid.New(),
		SenderID:  creator.ID,
		GroupID:   &group.ID,
		Content:   "Welcome to the team!",
		Type:      domain.MessageTypeGroup,
		CreatedAt: time.Now(),
	}
	msg2 := &domain.Message{
		ID:        uuid.New(),
		SenderID:  member1.ID,
		GroupID:   &group.ID,
		Content:   "Thanks for adding me!",
		Type:      domain.MessageTypeGroup,
		CreatedAt: time.Now().Add(1 * time.Second),
	}

	require.NoError(t, msgRepo.Create(msg1))
	require.NoError(t, msgRepo.Create(msg2))

	// List messages by group
	messages, err := msgRepo.ListByGroupID(group.ID, nil, 10)
	require.NoError(t, err)
	assert.Len(t, messages, 2)

	// Verify all members are in the group
	isMember, err := groupRepo.IsMember(group.ID, creator.ID)
	require.NoError(t, err)
	assert.True(t, isMember)

	isMember, err = groupRepo.IsMember(group.ID, member1.ID)
	require.NoError(t, err)
	assert.True(t, isMember)

	isMember, err = groupRepo.IsMember(group.ID, member2.ID)
	require.NoError(t, err)
	assert.True(t, isMember)
}

// TestIntegration_RefreshTokenWithUser tests the integration between
// User and RefreshToken repositories
func TestIntegration_RefreshTokenWithUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewRefreshTokenRepository(db)

	// Create user
	user := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, userRepo.Create(user))

	// Create refresh token for user
	token := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     "refresh-token-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
	require.NoError(t, tokenRepo.Create(token))

	// Retrieve token and verify it references the correct user
	foundToken, err := tokenRepo.GetByToken("refresh-token-123")
	require.NoError(t, err)
	assert.Equal(t, user.ID, foundToken.UserID)

	// Verify user can be retrieved
	foundUser, err := userRepo.GetByID(foundToken.UserID)
	require.NoError(t, err)
	assert.Equal(t, user.Username, foundUser.Username)

	// Delete all tokens for user
	err = tokenRepo.DeleteByUserID(user.ID)
	require.NoError(t, err)

	// Verify token is deleted
	_, err = tokenRepo.GetByToken("refresh-token-123")
	assert.ErrorIs(t, err, domain.ErrInvalidToken)
}

// TestIntegration_ConversationLookupByParticipants tests finding a conversation
// between two users
func TestIntegration_ConversationLookupByParticipants(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	convRepo := repository.NewConversationRepository(db)

	// Create users
	user1 := &domain.User{
		ID:           uuid.New(),
		Username:     "user1",
		Email:        "user1@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	user2 := &domain.User{
		ID:           uuid.New(),
		Username:     "user2",
		Email:        "user2@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	require.NoError(t, userRepo.Create(user1))
	require.NoError(t, userRepo.Create(user2))

	// Create conversation
	conv := &domain.Conversation{
		ID:           uuid.New(),
		Participants: []uuid.UUID{user1.ID, user2.ID},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, convRepo.Create(conv))

	// Find conversation by participants (both orderings)
	found1, err := convRepo.GetByParticipants(user1.ID, user2.ID)
	require.NoError(t, err)
	assert.Equal(t, conv.ID, found1.ID)

	found2, err := convRepo.GetByParticipants(user2.ID, user1.ID)
	require.NoError(t, err)
	assert.Equal(t, conv.ID, found2.ID)

	// List conversations for each user
	convs1, err := convRepo.ListByUserID(user1.ID)
	require.NoError(t, err)
	assert.Len(t, convs1, 1)
	assert.Equal(t, conv.ID, convs1[0].ID)

	convs2, err := convRepo.ListByUserID(user2.ID)
	require.NoError(t, err)
	assert.Len(t, convs2, 1)
	assert.Equal(t, conv.ID, convs2[0].ID)
}

// TestIntegration_MessagePagination tests cursor-based pagination across repositories
func TestIntegration_MessagePagination(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	defer testutil.TeardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	convRepo := repository.NewConversationRepository(db)
	msgRepo := repository.NewMessageRepository(db)

	// Create users
	user1 := &domain.User{
		ID:           uuid.New(),
		Username:     "sender",
		Email:        "sender@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	user2 := &domain.User{
		ID:           uuid.New(),
		Username:     "receiver",
		Email:        "receiver@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	require.NoError(t, userRepo.Create(user1))
	require.NoError(t, userRepo.Create(user2))

	// Create conversation
	conv := &domain.Conversation{
		ID:           uuid.New(),
		Participants: []uuid.UUID{user1.ID, user2.ID},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, convRepo.Create(conv))

	// Create 5 messages with specific timestamps
	now := time.Now()
	for i := 0; i < 5; i++ {
		msg := &domain.Message{
			ID:             uuid.New(),
			SenderID:       user1.ID,
			ConversationID: &conv.ID,
			Content:        "Message " + string(rune('A'+i)),
			Type:           domain.MessageTypeConversation,
			CreatedAt:      now.Add(time.Duration(i) * time.Minute),
		}
		require.NoError(t, msgRepo.Create(msg))
	}

	// Get first page (2 messages)
	page1, err := msgRepo.ListByConversationID(conv.ID, nil, 2)
	require.NoError(t, err)
	assert.Len(t, page1, 2)

	// Get second page using cursor from last message of page 1
	cursor := page1[1].CreatedAt
	page2, err := msgRepo.ListByConversationID(conv.ID, &cursor, 2)
	require.NoError(t, err)
	assert.Len(t, page2, 2)

	// Verify no overlap between pages
	for _, msg1 := range page1 {
		for _, msg2 := range page2 {
			assert.NotEqual(t, msg1.ID, msg2.ID)
		}
	}

	// Get third page
	cursor = page2[1].CreatedAt
	page3, err := msgRepo.ListByConversationID(conv.ID, &cursor, 2)
	require.NoError(t, err)
	assert.Len(t, page3, 1) // Only 1 message left
}

// ============================================================================
// Helper Functions
// ============================================================================

// Helper function to create test users
func createTestUser(t *testing.T, db *sql.DB, id uuid.UUID, username, email string) {
	t.Helper()
	query := `
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := db.Exec(query, id, username, email, "hash", time.Now(), time.Now())
	require.NoError(t, err)
}

// Helper function to create test conversation
func createTestConversation(t *testing.T, db *sql.DB, user1ID, user2ID uuid.UUID) uuid.UUID {
	t.Helper()
	convID := uuid.New()

	// Create conversation
	_, err := db.Exec(`
		INSERT INTO conversations (id, created_at, updated_at)
		VALUES ($1, $2, $3)
	`, convID, time.Now(), time.Now())
	require.NoError(t, err)

	// Add participants
	_, err = db.Exec(`
		INSERT INTO conversation_participants (conversation_id, user_id)
		VALUES ($1, $2), ($1, $3)
	`, convID, user1ID, user2ID)
	require.NoError(t, err)

	return convID
}

// Helper function to create test group
func createTestGroup(t *testing.T, db *sql.DB, name string, createdBy uuid.UUID, members []uuid.UUID) uuid.UUID {
	t.Helper()
	groupID := uuid.New()

	// Create group
	_, err := db.Exec(`
		INSERT INTO groups (id, name, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, groupID, name, createdBy, time.Now(), time.Now())
	require.NoError(t, err)

	// Add members
	for _, memberID := range members {
		_, err = db.Exec(`
			INSERT INTO group_members (group_id, user_id)
			VALUES ($1, $2)
		`, groupID, memberID)
		require.NoError(t, err)
	}

	return groupID
}
