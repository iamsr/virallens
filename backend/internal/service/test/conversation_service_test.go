package service_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/virallens/backend/internal/domain"
	"github.com/yourusername/virallens/backend/internal/service"
	"github.com/yourusername/virallens/backend/test/mocks"
)


func TestConversationService_CreateOrGet_CreatesNew(t *testing.T) {
	convRepo := mocks.NewMockConversationRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewConversationService(convRepo, userRepo)

	user1ID := uuid.New()
	user2ID := uuid.New()

	user1 := &domain.User{ID: user1ID, Username: "user1"}
	user2 := &domain.User{ID: user2ID, Username: "user2"}

	// Mock: both users exist
	userRepo.On("GetByID", user1ID).Return(user1, nil)
	userRepo.On("GetByID", user2ID).Return(user2, nil)

	// Mock: conversation doesn't exist
	convRepo.On("GetByParticipants", user1ID, user2ID).Return(nil, domain.ErrConversationNotFound)

	// Mock: create conversation
	convRepo.On("Create", mock.AnythingOfType("*domain.Conversation")).Return(nil)

	conv, err := svc.CreateOrGet(user1ID, user2ID)
	require.NoError(t, err)
	require.NotNil(t, conv)
	assert.Len(t, conv.Participants, 2)
	assert.Contains(t, conv.Participants, user1ID)
	assert.Contains(t, conv.Participants, user2ID)

	userRepo.AssertExpectations(t)
	convRepo.AssertExpectations(t)
}

func TestConversationService_CreateOrGet_ReturnsExisting(t *testing.T) {
	convRepo := mocks.NewMockConversationRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewConversationService(convRepo, userRepo)

	user1ID := uuid.New()
	user2ID := uuid.New()

	user1 := &domain.User{ID: user1ID, Username: "user1"}
	user2 := &domain.User{ID: user2ID, Username: "user2"}

	existingConv := &domain.Conversation{
		ID:           uuid.New(),
		Participants: []uuid.UUID{user1ID, user2ID},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Mock: both users exist
	userRepo.On("GetByID", user1ID).Return(user1, nil)
	userRepo.On("GetByID", user2ID).Return(user2, nil)

	// Mock: conversation already exists
	convRepo.On("GetByParticipants", user1ID, user2ID).Return(existingConv, nil)

	conv, err := svc.CreateOrGet(user1ID, user2ID)
	require.NoError(t, err)
	require.NotNil(t, conv)
	assert.Equal(t, existingConv.ID, conv.ID)

	userRepo.AssertExpectations(t)
	convRepo.AssertExpectations(t)
}

func TestConversationService_CreateOrGet_UserNotFound(t *testing.T) {
	convRepo := mocks.NewMockConversationRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewConversationService(convRepo, userRepo)

	user1ID := uuid.New()
	user2ID := uuid.New()

	// Mock: first user doesn't exist
	userRepo.On("GetByID", user1ID).Return(nil, domain.ErrUserNotFound)

	_, err := svc.CreateOrGet(user1ID, user2ID)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)

	userRepo.AssertExpectations(t)
}

func TestConversationService_GetByID_Success(t *testing.T) {
	convRepo := mocks.NewMockConversationRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewConversationService(convRepo, userRepo)

	userID := uuid.New()
	convID := uuid.New()

	conv := &domain.Conversation{
		ID:           convID,
		Participants: []uuid.UUID{userID, uuid.New()},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Mock: conversation exists
	convRepo.On("GetByID", convID).Return(conv, nil)

	result, err := svc.GetByID(convID, userID)
	require.NoError(t, err)
	assert.Equal(t, convID, result.ID)

	convRepo.AssertExpectations(t)
}

func TestConversationService_GetByID_NotParticipant(t *testing.T) {
	convRepo := mocks.NewMockConversationRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewConversationService(convRepo, userRepo)

	userID := uuid.New()
	convID := uuid.New()

	conv := &domain.Conversation{
		ID:           convID,
		Participants: []uuid.UUID{uuid.New(), uuid.New()}, // Different users
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Mock: conversation exists
	convRepo.On("GetByID", convID).Return(conv, nil)

	_, err := svc.GetByID(convID, userID)
	assert.ErrorIs(t, err, domain.ErrNotConversationMember)

	convRepo.AssertExpectations(t)
}

func TestConversationService_ListByUserID_Success(t *testing.T) {
	convRepo := mocks.NewMockConversationRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewConversationService(convRepo, userRepo)

	userID := uuid.New()
	user := &domain.User{ID: userID, Username: "user"}

	conversations := []*domain.Conversation{
		{
			ID:           uuid.New(),
			Participants: []uuid.UUID{userID, uuid.New()},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.New(),
			Participants: []uuid.UUID{userID, uuid.New()},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	// Mock: user exists
	userRepo.On("GetByID", userID).Return(user, nil)

	// Mock: list conversations
	convRepo.On("ListByUserID", userID).Return(conversations, nil)

	result, err := svc.ListByUserID(userID)
	require.NoError(t, err)
	assert.Len(t, result, 2)

	userRepo.AssertExpectations(t)
	convRepo.AssertExpectations(t)
}

func TestConversationService_AddParticipant_Success(t *testing.T) {
	convRepo := mocks.NewMockConversationRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewConversationService(convRepo, userRepo)

	convID := uuid.New()
	requestorID := uuid.New()
	newUserID := uuid.New()

	conv := &domain.Conversation{
		ID:           convID,
		Participants: []uuid.UUID{requestorID, uuid.New()},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	newUser := &domain.User{ID: newUserID, Username: "newuser"}

	// Mock: conversation exists
	convRepo.On("GetByID", convID).Return(conv, nil)

	// Mock: new user exists
	userRepo.On("GetByID", newUserID).Return(newUser, nil)

	// Mock: add participant
	convRepo.On("AddParticipant", convID, newUserID).Return(nil)

	err := svc.AddParticipant(convID, newUserID, requestorID)
	require.NoError(t, err)

	convRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestConversationService_AddParticipant_Unauthorized(t *testing.T) {
	convRepo := mocks.NewMockConversationRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewConversationService(convRepo, userRepo)

	convID := uuid.New()
	requestorID := uuid.New()
	newUserID := uuid.New()

	conv := &domain.Conversation{
		ID:           convID,
		Participants: []uuid.UUID{uuid.New(), uuid.New()}, // Requestor not in list
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Mock: conversation exists
	convRepo.On("GetByID", convID).Return(conv, nil)

	err := svc.AddParticipant(convID, newUserID, requestorID)
	assert.ErrorIs(t, err, domain.ErrUnauthorized)

	convRepo.AssertExpectations(t)
}
