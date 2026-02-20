package chat

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/virallens/backend/models"
	"github.com/yourusername/virallens/backend/modules/user"
)

var (
	ErrUnauthorized = errors.New("unauthorized access")
)

type ConversationService interface {
	CreateOrGet(user1ID, user2ID uuid.UUID) (*models.Conversation, error)
	GetByID(conversationID uuid.UUID) (*models.Conversation, error)
	ListUserConversations(userID uuid.UUID) ([]*models.Conversation, error)
}

type conversationSvc struct {
	repo     ConversationRepository
	userRepo user.Repository
}

func NewConversationService(repo ConversationRepository, userRepo user.Repository) ConversationService {
	return &conversationSvc{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (s *conversationSvc) CreateOrGet(user1ID, user2ID uuid.UUID) (*models.Conversation, error) {
	if user1ID == user2ID {
		return nil, errors.New("cannot create conversation with yourself")
	}

	_, err := s.userRepo.GetByID(user2ID)
	if err != nil {
		return nil, errors.New("other user not found")
	}

	existingConv, err := s.repo.GetByParticipants(user1ID, user2ID)
	if err != nil {
		return nil, err
	}
	if existingConv != nil {
		return existingConv, nil
	}

	conv := &models.Conversation{
		ID:           uuid.New(),
		Participant1: user1ID,
		Participant2: user2ID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.Create(conv); err != nil {
		return nil, err
	}

	return conv, nil
}

func (s *conversationSvc) GetByID(conversationID uuid.UUID) (*models.Conversation, error) {
	return s.repo.GetByID(conversationID)
}

func (s *conversationSvc) ListUserConversations(userID uuid.UUID) ([]*models.Conversation, error) {
	return s.repo.ListByUserID(userID)
}
