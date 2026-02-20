package chat

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/iamsr/virallens/backend/models"
	"github.com/iamsr/virallens/backend/modules/user"
)

type MessageService interface {
	SendConversationMessage(senderID, conversationID uuid.UUID, content string) (*models.Message, error)
	SendGroupMessage(senderID, groupID uuid.UUID, content string) (*models.Message, error)
	GetConversationMessages(userID, conversationID uuid.UUID, cursor *time.Time, limit int) ([]*models.Message, error)
	GetGroupMessages(userID, groupID uuid.UUID, cursor *time.Time, limit int) ([]*models.Message, error)
}

type messageSvc struct {
	messageRepo      MessageRepository
	conversationRepo ConversationRepository
	groupRepo        GroupRepository
	userRepo         user.Repository
}

func NewMessageService(
	messageRepo MessageRepository,
	conversationRepo ConversationRepository,
	groupRepo GroupRepository,
	userRepo user.Repository,
) MessageService {
	return &messageSvc{
		messageRepo:      messageRepo,
		conversationRepo: conversationRepo,
		groupRepo:        groupRepo,
		userRepo:         userRepo,
	}
}

func normalizeLimit(limit int) int {
	if limit <= 0 || limit > 100 {
		return 50
	}
	return limit
}

func (s *messageSvc) SendConversationMessage(senderID, conversationID uuid.UUID, content string) (*models.Message, error) {
	if content == "" {
		return nil, errors.New("message content cannot be empty")
	}

	_, err := s.userRepo.GetByID(senderID)
	if err != nil {
		return nil, err
	}

	_, err = s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return nil, err
	}

	isParticipant, err := s.conversationRepo.IsParticipant(conversationID, senderID)
	if err != nil || !isParticipant {
		return nil, ErrUnauthorized
	}

	message := &models.Message{
		ID:             uuid.New(),
		SenderID:       senderID,
		ConversationID: &conversationID,
		Content:        content,
		Type:           models.MessageTypeConversation,
		CreatedAt:      time.Now(),
	}

	if err := s.messageRepo.Create(message); err != nil {
		return nil, err
	}

	return message, nil
}

func (s *messageSvc) SendGroupMessage(senderID, groupID uuid.UUID, content string) (*models.Message, error) {
	if content == "" {
		return nil, errors.New("message content cannot be empty")
	}

	_, err := s.userRepo.GetByID(senderID)
	if err != nil {
		return nil, err
	}

	_, err = s.groupRepo.GetByID(groupID)
	if err != nil {
		return nil, err
	}

	isMember, err := s.groupRepo.IsMember(groupID, senderID)
	if err != nil || !isMember {
		return nil, ErrUnauthorized
	}

	message := &models.Message{
		ID:        uuid.New(),
		SenderID:  senderID,
		GroupID:   &groupID,
		Content:   content,
		Type:      models.MessageTypeGroup,
		CreatedAt: time.Now(),
	}

	if err := s.messageRepo.Create(message); err != nil {
		return nil, err
	}

	return message, nil
}

func (s *messageSvc) GetConversationMessages(userID, conversationID uuid.UUID, cursor *time.Time, limit int) ([]*models.Message, error) {
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	_, err = s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return nil, err
	}

	isParticipant, err := s.conversationRepo.IsParticipant(conversationID, userID)
	if err != nil || !isParticipant {
		return nil, ErrUnauthorized
	}

	limit = normalizeLimit(limit)
	return s.messageRepo.ListByConversationID(conversationID, cursor, limit)
}

func (s *messageSvc) GetGroupMessages(userID, groupID uuid.UUID, cursor *time.Time, limit int) ([]*models.Message, error) {
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	_, err = s.groupRepo.GetByID(groupID)
	if err != nil {
		return nil, err
	}

	isMember, err := s.groupRepo.IsMember(groupID, userID)
	if err != nil || !isMember {
		return nil, ErrUnauthorized
	}

	limit = normalizeLimit(limit)
	return s.messageRepo.ListByGroupID(groupID, cursor, limit)
}
