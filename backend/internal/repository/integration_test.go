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

// TestIntegration_ConversationWithMessages tests the integration between
// User, Conversation, and Message repositories
func TestIntegration_ConversationWithMessages(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

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
	db := setupTestDB(t)
	defer db.Close()

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
	db := setupTestDB(t)
	defer db.Close()

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
	db := setupTestDB(t)
	defer db.Close()

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
	db := setupTestDB(t)
	defer db.Close()

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
