package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourusername/virallens/backend/internal/domain"
)

// MockMessageRepository is a mock implementation of MessageRepository
type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) Create(message *domain.Message) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockMessageRepository) GetByID(id uuid.UUID) (*domain.Message, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Message), args.Error(1)
}

func (m *MockMessageRepository) ListByConversationID(conversationID uuid.UUID, cursor *time.Time, limit int) ([]*domain.Message, error) {
	args := m.Called(conversationID, cursor, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Message), args.Error(1)
}

func (m *MockMessageRepository) ListByGroupID(groupID uuid.UUID, cursor *time.Time, limit int) ([]*domain.Message, error) {
	args := m.Called(groupID, cursor, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Message), args.Error(1)
}

// Tests for SendConversationMessage

func TestMessageService_SendConversationMessage_Success(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockConvRepo := new(MockConversationRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewMessageService(mockMessageRepo, mockConvRepo, mockUserRepo, nil)

	senderID := uuid.New()
	conversationID := uuid.New()
	content := "Hello!"

	// Mock: user exists
	mockUserRepo.On("GetByID", senderID).Return(&domain.User{ID: senderID}, nil)

	// Mock: conversation exists and user is participant
	mockConvRepo.On("GetByID", conversationID).Return(&domain.Conversation{ID: conversationID}, nil)
	mockConvRepo.On("IsParticipant", conversationID, senderID).Return(true, nil)

	// Mock: message creation
	mockMessageRepo.On("Create", mock.MatchedBy(func(msg *domain.Message) bool {
		return msg.SenderID == senderID &&
			msg.ConversationID != nil &&
			*msg.ConversationID == conversationID &&
			msg.Content == content &&
			msg.Type == domain.MessageTypeConversation
	})).Return(nil)

	message, err := service.SendConversationMessage(senderID, conversationID, content)

	assert.NoError(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, senderID, message.SenderID)
	assert.Equal(t, conversationID, *message.ConversationID)
	assert.Equal(t, content, message.Content)
	assert.Equal(t, domain.MessageTypeConversation, message.Type)

	mockUserRepo.AssertExpectations(t)
	mockConvRepo.AssertExpectations(t)
	mockMessageRepo.AssertExpectations(t)
}

func TestMessageService_SendConversationMessage_NotParticipant(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockConvRepo := new(MockConversationRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewMessageService(mockMessageRepo, mockConvRepo, mockUserRepo, nil)

	senderID := uuid.New()
	conversationID := uuid.New()
	content := "Hello!"

	// Mock: user exists
	mockUserRepo.On("GetByID", senderID).Return(&domain.User{ID: senderID}, nil)

	// Mock: conversation exists but user is NOT participant
	mockConvRepo.On("GetByID", conversationID).Return(&domain.Conversation{ID: conversationID}, nil)
	mockConvRepo.On("IsParticipant", conversationID, senderID).Return(false, nil)

	message, err := service.SendConversationMessage(senderID, conversationID, content)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrUnauthorized, err)
	assert.Nil(t, message)

	mockUserRepo.AssertExpectations(t)
	mockConvRepo.AssertExpectations(t)
	mockMessageRepo.AssertNotCalled(t, "Create")
}

func TestMessageService_SendConversationMessage_EmptyContent(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockConvRepo := new(MockConversationRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewMessageService(mockMessageRepo, mockConvRepo, mockUserRepo, nil)

	senderID := uuid.New()
	conversationID := uuid.New()
	content := ""

	message, err := service.SendConversationMessage(senderID, conversationID, content)

	assert.Error(t, err)
	assert.Nil(t, message)

	mockUserRepo.AssertNotCalled(t, "GetByID")
	mockConvRepo.AssertNotCalled(t, "GetByID")
	mockMessageRepo.AssertNotCalled(t, "Create")
}

// Tests for SendGroupMessage

func TestMessageService_SendGroupMessage_Success(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockGroupRepo := new(MockGroupRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewMessageService(mockMessageRepo, nil, mockUserRepo, mockGroupRepo)

	senderID := uuid.New()
	groupID := uuid.New()
	content := "Hello group!"

	// Mock: user exists
	mockUserRepo.On("GetByID", senderID).Return(&domain.User{ID: senderID}, nil)

	// Mock: group exists and user is member
	mockGroupRepo.On("GetByID", groupID).Return(&domain.Group{ID: groupID}, nil)
	mockGroupRepo.On("IsMember", groupID, senderID).Return(true, nil)

	// Mock: message creation
	mockMessageRepo.On("Create", mock.MatchedBy(func(msg *domain.Message) bool {
		return msg.SenderID == senderID &&
			msg.GroupID != nil &&
			*msg.GroupID == groupID &&
			msg.Content == content &&
			msg.Type == domain.MessageTypeGroup
	})).Return(nil)

	message, err := service.SendGroupMessage(senderID, groupID, content)

	assert.NoError(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, senderID, message.SenderID)
	assert.Equal(t, groupID, *message.GroupID)
	assert.Equal(t, content, message.Content)
	assert.Equal(t, domain.MessageTypeGroup, message.Type)

	mockUserRepo.AssertExpectations(t)
	mockGroupRepo.AssertExpectations(t)
	mockMessageRepo.AssertExpectations(t)
}

func TestMessageService_SendGroupMessage_NotMember(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockGroupRepo := new(MockGroupRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewMessageService(mockMessageRepo, nil, mockUserRepo, mockGroupRepo)

	senderID := uuid.New()
	groupID := uuid.New()
	content := "Hello group!"

	// Mock: user exists
	mockUserRepo.On("GetByID", senderID).Return(&domain.User{ID: senderID}, nil)

	// Mock: group exists but user is NOT member
	mockGroupRepo.On("GetByID", groupID).Return(&domain.Group{ID: groupID}, nil)
	mockGroupRepo.On("IsMember", groupID, senderID).Return(false, nil)

	message, err := service.SendGroupMessage(senderID, groupID, content)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrUnauthorized, err)
	assert.Nil(t, message)

	mockUserRepo.AssertExpectations(t)
	mockGroupRepo.AssertExpectations(t)
	mockMessageRepo.AssertNotCalled(t, "Create")
}

// Tests for GetConversationMessages

func TestMessageService_GetConversationMessages_Success(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockConvRepo := new(MockConversationRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewMessageService(mockMessageRepo, mockConvRepo, mockUserRepo, nil)

	userID := uuid.New()
	conversationID := uuid.New()
	limit := 20

	// Mock: user exists
	mockUserRepo.On("GetByID", userID).Return(&domain.User{ID: userID}, nil)

	// Mock: conversation exists and user is participant
	mockConvRepo.On("GetByID", conversationID).Return(&domain.Conversation{ID: conversationID}, nil)
	mockConvRepo.On("IsParticipant", conversationID, userID).Return(true, nil)

	// Mock: messages retrieved
	expectedMessages := []*domain.Message{
		{ID: uuid.New(), Content: "Message 1"},
		{ID: uuid.New(), Content: "Message 2"},
	}
	mockMessageRepo.On("ListByConversationID", conversationID, (*time.Time)(nil), limit).Return(expectedMessages, nil)

	messages, err := service.GetConversationMessages(userID, conversationID, nil, limit)

	assert.NoError(t, err)
	assert.NotNil(t, messages)
	assert.Len(t, messages, 2)

	mockUserRepo.AssertExpectations(t)
	mockConvRepo.AssertExpectations(t)
	mockMessageRepo.AssertExpectations(t)
}

func TestMessageService_GetConversationMessages_NotParticipant(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockConvRepo := new(MockConversationRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewMessageService(mockMessageRepo, mockConvRepo, mockUserRepo, nil)

	userID := uuid.New()
	conversationID := uuid.New()

	// Mock: user exists
	mockUserRepo.On("GetByID", userID).Return(&domain.User{ID: userID}, nil)

	// Mock: conversation exists but user is NOT participant
	mockConvRepo.On("GetByID", conversationID).Return(&domain.Conversation{ID: conversationID}, nil)
	mockConvRepo.On("IsParticipant", conversationID, userID).Return(false, nil)

	messages, err := service.GetConversationMessages(userID, conversationID, nil, 20)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrUnauthorized, err)
	assert.Nil(t, messages)

	mockUserRepo.AssertExpectations(t)
	mockConvRepo.AssertExpectations(t)
	mockMessageRepo.AssertNotCalled(t, "ListByConversationID")
}

// Tests for GetGroupMessages

func TestMessageService_GetGroupMessages_Success(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockGroupRepo := new(MockGroupRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewMessageService(mockMessageRepo, nil, mockUserRepo, mockGroupRepo)

	userID := uuid.New()
	groupID := uuid.New()
	limit := 20

	// Mock: user exists
	mockUserRepo.On("GetByID", userID).Return(&domain.User{ID: userID}, nil)

	// Mock: group exists and user is member
	mockGroupRepo.On("GetByID", groupID).Return(&domain.Group{ID: groupID}, nil)
	mockGroupRepo.On("IsMember", groupID, userID).Return(true, nil)

	// Mock: messages retrieved
	expectedMessages := []*domain.Message{
		{ID: uuid.New(), Content: "Group message 1"},
		{ID: uuid.New(), Content: "Group message 2"},
	}
	mockMessageRepo.On("ListByGroupID", groupID, (*time.Time)(nil), limit).Return(expectedMessages, nil)

	messages, err := service.GetGroupMessages(userID, groupID, nil, limit)

	assert.NoError(t, err)
	assert.NotNil(t, messages)
	assert.Len(t, messages, 2)

	mockUserRepo.AssertExpectations(t)
	mockGroupRepo.AssertExpectations(t)
	mockMessageRepo.AssertExpectations(t)
}

func TestMessageService_GetGroupMessages_NotMember(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockGroupRepo := new(MockGroupRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewMessageService(mockMessageRepo, nil, mockUserRepo, mockGroupRepo)

	userID := uuid.New()
	groupID := uuid.New()

	// Mock: user exists
	mockUserRepo.On("GetByID", userID).Return(&domain.User{ID: userID}, nil)

	// Mock: group exists but user is NOT member
	mockGroupRepo.On("GetByID", groupID).Return(&domain.Group{ID: groupID}, nil)
	mockGroupRepo.On("IsMember", groupID, userID).Return(false, nil)

	messages, err := service.GetGroupMessages(userID, groupID, nil, 20)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrUnauthorized, err)
	assert.Nil(t, messages)

	mockUserRepo.AssertExpectations(t)
	mockGroupRepo.AssertExpectations(t)
	mockMessageRepo.AssertNotCalled(t, "ListByGroupID")
}
