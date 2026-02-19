package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/virallens/backend/internal/domain"
)

type ConversationService interface {
	CreateOrGet(user1ID, user2ID uuid.UUID) (*domain.Conversation, error)
	GetByID(id uuid.UUID, userID uuid.UUID) (*domain.Conversation, error)
	ListByUserID(userID uuid.UUID) ([]*domain.Conversation, error)
	AddParticipant(conversationID, userID, requestorID uuid.UUID) error
}

type conversationService struct {
	conversationRepo domain.ConversationRepository
	userRepo         domain.UserRepository
}

// NewConversationService creates a new conversation service
func NewConversationService(
	conversationRepo domain.ConversationRepository,
	userRepo domain.UserRepository,
) ConversationService {
	return &conversationService{
		conversationRepo: conversationRepo,
		userRepo:         userRepo,
	}
}

func (s *conversationService) CreateOrGet(user1ID, user2ID uuid.UUID) (*domain.Conversation, error) {
	// Validate both users exist
	_, err := s.userRepo.GetByID(user1ID)
	if err != nil {
		return nil, err
	}

	_, err = s.userRepo.GetByID(user2ID)
	if err != nil {
		return nil, err
	}

	// Check if conversation already exists
	conv, err := s.conversationRepo.GetByParticipants(user1ID, user2ID)
	if err == nil {
		return conv, nil
	}

	// Create new conversation if not found
	if err == domain.ErrConversationNotFound {
		conv := &domain.Conversation{
			ID:           uuid.New(),
			Participants: []uuid.UUID{user1ID, user2ID},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := s.conversationRepo.Create(conv); err != nil {
			return nil, err
		}

		return conv, nil
	}

	return nil, err
}

func (s *conversationService) GetByID(id uuid.UUID, userID uuid.UUID) (*domain.Conversation, error) {
	conv, err := s.conversationRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Verify user is a participant
	isParticipant := false
	for _, participantID := range conv.Participants {
		if participantID == userID {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		return nil, domain.ErrNotConversationMember
	}

	return conv, nil
}

func (s *conversationService) ListByUserID(userID uuid.UUID) ([]*domain.Conversation, error) {
	// Verify user exists
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	return s.conversationRepo.ListByUserID(userID)
}

func (s *conversationService) AddParticipant(conversationID, userID, requestorID uuid.UUID) error {
	// Get conversation
	conv, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return err
	}

	// Verify requestor is a participant
	isRequestorParticipant := false
	for _, participantID := range conv.Participants {
		if participantID == requestorID {
			isRequestorParticipant = true
			break
		}
	}

	if !isRequestorParticipant {
		return domain.ErrUnauthorized
	}

	// Verify user to add exists
	_, err = s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// Add participant
	return s.conversationRepo.AddParticipant(conversationID, userID)
}
