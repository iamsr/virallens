package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/virallens/backend/internal/domain"
)

type MessageService struct {
	messageRepo      domain.MessageRepository
	conversationRepo domain.ConversationRepository
	userRepo         domain.UserRepository
	groupRepo        domain.GroupRepository
}

func NewMessageService(
	messageRepo domain.MessageRepository,
	conversationRepo domain.ConversationRepository,
	userRepo domain.UserRepository,
	groupRepo domain.GroupRepository,
) *MessageService {
	return &MessageService{
		messageRepo:      messageRepo,
		conversationRepo: conversationRepo,
		userRepo:         userRepo,
		groupRepo:        groupRepo,
	}
}

// SendConversationMessage sends a message to a conversation
func (s *MessageService) SendConversationMessage(senderID, conversationID uuid.UUID, content string) (*domain.Message, error) {
	// Validate content
	if content == "" {
		return nil, errors.New("message content cannot be empty")
	}

	// Check if sender exists
	_, err := s.userRepo.GetByID(senderID)
	if err != nil {
		return nil, err
	}

	// Check if conversation exists
	_, err = s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return nil, err
	}

	// Check if sender is participant
	isParticipant, err := s.conversationRepo.IsParticipant(conversationID, senderID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, domain.ErrUnauthorized
	}

	// Create message
	message := &domain.Message{
		ID:             uuid.New(),
		SenderID:       senderID,
		ConversationID: &conversationID,
		Content:        content,
		Type:           domain.MessageTypeConversation,
		CreatedAt:      time.Now(),
	}

	err = s.messageRepo.Create(message)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// SendGroupMessage sends a message to a group
func (s *MessageService) SendGroupMessage(senderID, groupID uuid.UUID, content string) (*domain.Message, error) {
	// Validate content
	if content == "" {
		return nil, errors.New("message content cannot be empty")
	}

	// Check if sender exists
	_, err := s.userRepo.GetByID(senderID)
	if err != nil {
		return nil, err
	}

	// Check if group exists
	_, err = s.groupRepo.GetByID(groupID)
	if err != nil {
		return nil, err
	}

	// Check if sender is member
	isMember, err := s.groupRepo.IsMember(groupID, senderID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrUnauthorized
	}

	// Create message
	message := &domain.Message{
		ID:        uuid.New(),
		SenderID:  senderID,
		GroupID:   &groupID,
		Content:   content,
		Type:      domain.MessageTypeGroup,
		CreatedAt: time.Now(),
	}

	err = s.messageRepo.Create(message)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// GetConversationMessages retrieves messages from a conversation with cursor-based pagination
func (s *MessageService) GetConversationMessages(userID, conversationID uuid.UUID, cursor *time.Time, limit int) ([]*domain.Message, error) {
	// Check if user exists
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Check if conversation exists
	_, err = s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return nil, err
	}

	// Check if user is participant
	isParticipant, err := s.conversationRepo.IsParticipant(conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, domain.ErrUnauthorized
	}

	// Retrieve messages
	messages, err := s.messageRepo.ListByConversationID(conversationID, cursor, limit)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

// GetGroupMessages retrieves messages from a group with cursor-based pagination
func (s *MessageService) GetGroupMessages(userID, groupID uuid.UUID, cursor *time.Time, limit int) ([]*domain.Message, error) {
	// Check if user exists
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Check if group exists
	_, err = s.groupRepo.GetByID(groupID)
	if err != nil {
		return nil, err
	}

	// Check if user is member
	isMember, err := s.groupRepo.IsMember(groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrUnauthorized
	}

	// Retrieve messages
	messages, err := s.messageRepo.ListByGroupID(groupID, cursor, limit)
	if err != nil {
		return nil, err
	}

	return messages, nil
}
