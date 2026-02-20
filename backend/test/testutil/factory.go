package testutil

import (
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/virallens/backend/internal/domain"
)

// CreateTestUser creates a test user with default values
func CreateTestUser() *domain.User {
	return &domain.User{
		ID:           uuid.New(),
		Username:     "testuser_" + uuid.New().String()[:8],
		Email:        "test_" + uuid.New().String()[:8] + "@example.com",
		PasswordHash: "$2a$10$abcdefghijklmnopqrstuv", // bcrypt hash
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// CreateTestConversation creates a test conversation
func CreateTestConversation(participants ...uuid.UUID) *domain.Conversation {
	return &domain.Conversation{
		ID:           uuid.New(),
		Participants: participants,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// CreateTestGroup creates a test group
func CreateTestGroup(creatorID uuid.UUID) *domain.Group {
	return &domain.Group{
		ID:        uuid.New(),
		Name:      "Test Group " + uuid.New().String()[:8],
		CreatedBy: creatorID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateTestMessage creates a test message
func CreateTestMessage(senderID, conversationID uuid.UUID) *domain.Message {
	return &domain.Message{
		ID:             uuid.New(),
		SenderID:       senderID,
		ConversationID: &conversationID,
		GroupID:        nil,
		Content:        "Test message content",
		Type:           domain.MessageTypeConversation,
		CreatedAt:      time.Now(),
	}
}
