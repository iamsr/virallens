package unit

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
	"golang.org/x/crypto/bcrypt"
)

// MockJWTService - JWT service is not a domain interface, so we mock it manually
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateAccessToken(userID uuid.UUID) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) GenerateRefreshToken() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) ValidateAccessToken(tokenString string) (*service.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.Claims), args.Error(1)
}

// =============================================================================
// Auth Service Tests
// =============================================================================

func TestAuthService_Register_Success(t *testing.T) {
	userRepo := mocks.NewMockUserRepository(t)
	tokenRepo := mocks.NewMockRefreshTokenRepository(t)
	jwtService := new(MockJWTService)

	authService := service.NewAuthService(userRepo, tokenRepo, jwtService)

	req := &service.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	// Mock: user doesn't exist
	userRepo.On("GetByUsername", req.Username).Return(nil, domain.ErrUserNotFound)
	userRepo.On("GetByEmail", req.Email).Return(nil, domain.ErrUserNotFound)

	// Mock: user creation succeeds
	userRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)

	// Mock: token generation
	jwtService.On("GenerateAccessToken", mock.AnythingOfType("uuid.UUID")).Return("access-token", nil)
	jwtService.On("GenerateRefreshToken").Return("refresh-token", nil)

	// Mock: refresh token storage
	tokenRepo.On("Create", mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

	resp, err := authService.Register(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, req.Username, resp.User.Username)
	assert.Equal(t, req.Email, resp.User.Email)
	assert.Equal(t, "access-token", resp.AccessToken)
	assert.Equal(t, "refresh-token", resp.RefreshToken)

	// Verify password was hashed
	err = bcrypt.CompareHashAndPassword([]byte(resp.User.PasswordHash), []byte(req.Password))
	assert.NoError(t, err)

	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	jwtService.AssertExpectations(t)
}

func TestAuthService_Register_UserAlreadyExists(t *testing.T) {
	userRepo := mocks.NewMockUserRepository(t)
	tokenRepo := mocks.NewMockRefreshTokenRepository(t)
	jwtService := new(MockJWTService)

	authService := service.NewAuthService(userRepo, tokenRepo, jwtService)

	req := &service.RegisterRequest{
		Username: "existinguser",
		Email:    "existing@example.com",
		Password: "password123",
	}

	existingUser := &domain.User{
		ID:       uuid.New(),
		Username: "existinguser",
		Email:    "existing@example.com",
	}

	// Mock: user already exists
	userRepo.On("GetByUsername", req.Username).Return(existingUser, nil)

	_, err := authService.Register(req)
	assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)

	userRepo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	userRepo := mocks.NewMockUserRepository(t)
	tokenRepo := mocks.NewMockRefreshTokenRepository(t)
	jwtService := new(MockJWTService)

	authService := service.NewAuthService(userRepo, tokenRepo, jwtService)

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
	}

	req := &service.LoginRequest{
		Username: "testuser",
		Password: password,
	}

	// Mock: user exists
	userRepo.On("GetByUsername", req.Username).Return(user, nil)

	// Mock: delete old tokens
	tokenRepo.On("DeleteByUserID", user.ID).Return(nil)

	// Mock: token generation
	jwtService.On("GenerateAccessToken", user.ID).Return("access-token", nil)
	jwtService.On("GenerateRefreshToken").Return("refresh-token", nil)

	// Mock: refresh token storage
	tokenRepo.On("Create", mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

	resp, err := authService.Login(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, user.Username, resp.User.Username)
	assert.Equal(t, "access-token", resp.AccessToken)
	assert.Equal(t, "refresh-token", resp.RefreshToken)

	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	jwtService.AssertExpectations(t)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	userRepo := mocks.NewMockUserRepository(t)
	tokenRepo := mocks.NewMockRefreshTokenRepository(t)
	jwtService := new(MockJWTService)

	authService := service.NewAuthService(userRepo, tokenRepo, jwtService)

	req := &service.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	// Mock: user not found
	userRepo.On("GetByUsername", req.Username).Return(nil, domain.ErrUserNotFound)

	_, err := authService.Login(req)
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)

	userRepo.AssertExpectations(t)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	userRepo := mocks.NewMockUserRepository(t)
	tokenRepo := mocks.NewMockRefreshTokenRepository(t)
	jwtService := new(MockJWTService)

	authService := service.NewAuthService(userRepo, tokenRepo, jwtService)

	correctPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		PasswordHash: string(hashedPassword),
	}

	req := &service.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	// Mock: user exists
	userRepo.On("GetByUsername", req.Username).Return(user, nil)

	_, err := authService.Login(req)
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)

	userRepo.AssertExpectations(t)
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	userRepo := mocks.NewMockUserRepository(t)
	tokenRepo := mocks.NewMockRefreshTokenRepository(t)
	jwtService := new(MockJWTService)

	authService := service.NewAuthService(userRepo, tokenRepo, jwtService)

	userID := uuid.New()
	user := &domain.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
	}

	refreshToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     "refresh-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	// Mock: refresh token exists and is valid
	tokenRepo.On("GetByToken", "refresh-token").Return(refreshToken, nil)

	// Mock: user exists
	userRepo.On("GetByID", userID).Return(user, nil)

	// Mock: delete old token
	tokenRepo.On("DeleteByUserID", userID).Return(nil)

	// Mock: new token generation
	jwtService.On("GenerateAccessToken", userID).Return("new-access-token", nil)
	jwtService.On("GenerateRefreshToken").Return("new-refresh-token", nil)

	// Mock: new refresh token storage
	tokenRepo.On("Create", mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

	resp, err := authService.RefreshToken("refresh-token")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, user.Username, resp.User.Username)
	assert.Equal(t, "new-access-token", resp.AccessToken)
	assert.Equal(t, "new-refresh-token", resp.RefreshToken)

	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	jwtService.AssertExpectations(t)
}

func TestAuthService_RefreshToken_Expired(t *testing.T) {
	userRepo := mocks.NewMockUserRepository(t)
	tokenRepo := mocks.NewMockRefreshTokenRepository(t)
	jwtService := new(MockJWTService)

	authService := service.NewAuthService(userRepo, tokenRepo, jwtService)

	userID := uuid.New()
	refreshToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	// Mock: refresh token exists but is expired
	tokenRepo.On("GetByToken", "expired-token").Return(refreshToken, nil)
	tokenRepo.On("DeleteByUserID", userID).Return(nil)

	_, err := authService.RefreshToken("expired-token")
	assert.ErrorIs(t, err, domain.ErrTokenExpired)

	tokenRepo.AssertExpectations(t)
}

func TestAuthService_Logout(t *testing.T) {
	userRepo := mocks.NewMockUserRepository(t)
	tokenRepo := mocks.NewMockRefreshTokenRepository(t)
	jwtService := new(MockJWTService)

	authService := service.NewAuthService(userRepo, tokenRepo, jwtService)

	userID := uuid.New()

	// Mock: delete tokens
	tokenRepo.On("DeleteByUserID", userID).Return(nil)

	err := authService.Logout(userID)
	require.NoError(t, err)

	tokenRepo.AssertExpectations(t)
}

// =============================================================================
// Conversation Service Tests
// =============================================================================

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

// =============================================================================
// Group Service Tests
// =============================================================================

func TestGroupService_Create_Success(t *testing.T) {
	groupRepo := mocks.NewMockGroupRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewGroupService(groupRepo, userRepo)

	creatorID := uuid.New()
	member1ID := uuid.New()
	member2ID := uuid.New()

	creator := &domain.User{ID: creatorID, Username: "creator"}
	member1 := &domain.User{ID: member1ID, Username: "member1"}
	member2 := &domain.User{ID: member2ID, Username: "member2"}

	// Mock: all users exist
	userRepo.On("GetByID", creatorID).Return(creator, nil)
	userRepo.On("GetByID", member1ID).Return(member1, nil)
	userRepo.On("GetByID", member2ID).Return(member2, nil)

	// Mock: create group
	groupRepo.On("Create", mock.AnythingOfType("*domain.Group")).Return(nil)

	group, err := svc.Create("Test Group", creatorID, []uuid.UUID{member1ID, member2ID})
	require.NoError(t, err)
	require.NotNil(t, group)
	assert.Equal(t, "Test Group", group.Name)
	assert.Equal(t, creatorID, group.CreatedBy)
	assert.Len(t, group.Members, 3) // Creator auto-added
	assert.Contains(t, group.Members, creatorID)
	assert.Contains(t, group.Members, member1ID)
	assert.Contains(t, group.Members, member2ID)

	userRepo.AssertExpectations(t)
	groupRepo.AssertExpectations(t)
}

func TestGroupService_Create_CreatorInMemberList(t *testing.T) {
	groupRepo := mocks.NewMockGroupRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewGroupService(groupRepo, userRepo)

	creatorID := uuid.New()
	member1ID := uuid.New()

	creator := &domain.User{ID: creatorID, Username: "creator"}
	member1 := &domain.User{ID: member1ID, Username: "member1"}

	// Mock: all users exist
	userRepo.On("GetByID", creatorID).Return(creator, nil)
	userRepo.On("GetByID", member1ID).Return(member1, nil)

	// Mock: create group
	groupRepo.On("Create", mock.AnythingOfType("*domain.Group")).Return(nil)

	// Creator already in member list
	group, err := svc.Create("Test Group", creatorID, []uuid.UUID{creatorID, member1ID})
	require.NoError(t, err)
	assert.Len(t, group.Members, 2) // No duplicate

	userRepo.AssertExpectations(t)
	groupRepo.AssertExpectations(t)
}

func TestGroupService_Create_MemberNotFound(t *testing.T) {
	groupRepo := mocks.NewMockGroupRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewGroupService(groupRepo, userRepo)

	creatorID := uuid.New()
	invalidMemberID := uuid.New()

	creator := &domain.User{ID: creatorID, Username: "creator"}

	// Mock: creator exists, member doesn't
	userRepo.On("GetByID", creatorID).Return(creator, nil)
	userRepo.On("GetByID", invalidMemberID).Return(nil, domain.ErrUserNotFound)

	_, err := svc.Create("Test Group", creatorID, []uuid.UUID{invalidMemberID})
	assert.ErrorIs(t, err, domain.ErrUserNotFound)

	userRepo.AssertExpectations(t)
}

func TestGroupService_GetByID_Success(t *testing.T) {
	groupRepo := mocks.NewMockGroupRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewGroupService(groupRepo, userRepo)

	userID := uuid.New()
	groupID := uuid.New()

	group := &domain.Group{
		ID:        groupID,
		Name:      "Test Group",
		CreatedBy: userID,
		Members:   []uuid.UUID{userID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Mock: group exists
	groupRepo.On("GetByID", groupID).Return(group, nil)

	// Mock: user is member
	groupRepo.On("IsMember", groupID, userID).Return(true, nil)

	result, err := svc.GetByID(groupID, userID)
	require.NoError(t, err)
	assert.Equal(t, groupID, result.ID)

	groupRepo.AssertExpectations(t)
}

func TestGroupService_GetByID_NotMember(t *testing.T) {
	groupRepo := mocks.NewMockGroupRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewGroupService(groupRepo, userRepo)

	userID := uuid.New()
	groupID := uuid.New()

	group := &domain.Group{
		ID:        groupID,
		Name:      "Test Group",
		CreatedBy: uuid.New(),
		Members:   []uuid.UUID{uuid.New()},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Mock: group exists
	groupRepo.On("GetByID", groupID).Return(group, nil)

	// Mock: user is not a member
	groupRepo.On("IsMember", groupID, userID).Return(false, nil)

	_, err := svc.GetByID(groupID, userID)
	assert.ErrorIs(t, err, domain.ErrNotGroupMember)

	groupRepo.AssertExpectations(t)
}

func TestGroupService_ListByUserID_Success(t *testing.T) {
	groupRepo := mocks.NewMockGroupRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewGroupService(groupRepo, userRepo)

	userID := uuid.New()
	user := &domain.User{ID: userID, Username: "user"}

	groups := []*domain.Group{
		{
			ID:        uuid.New(),
			Name:      "Group 1",
			CreatedBy: userID,
			Members:   []uuid.UUID{userID},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "Group 2",
			CreatedBy: uuid.New(),
			Members:   []uuid.UUID{userID, uuid.New()},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Mock: user exists
	userRepo.On("GetByID", userID).Return(user, nil)

	// Mock: list groups
	groupRepo.On("ListByUserID", userID).Return(groups, nil)

	result, err := svc.ListByUserID(userID)
	require.NoError(t, err)
	assert.Len(t, result, 2)

	userRepo.AssertExpectations(t)
	groupRepo.AssertExpectations(t)
}

func TestGroupService_AddMember_Success(t *testing.T) {
	groupRepo := mocks.NewMockGroupRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewGroupService(groupRepo, userRepo)

	groupID := uuid.New()
	requestorID := uuid.New()
	newUserID := uuid.New()

	group := &domain.Group{
		ID:        groupID,
		Name:      "Test Group",
		CreatedBy: requestorID,
		Members:   []uuid.UUID{requestorID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	newUser := &domain.User{ID: newUserID, Username: "newuser"}

	// Mock: group exists
	groupRepo.On("GetByID", groupID).Return(group, nil)

	// Mock: requestor is member
	groupRepo.On("IsMember", groupID, requestorID).Return(true, nil)

	// Mock: new user exists
	userRepo.On("GetByID", newUserID).Return(newUser, nil)

	// Mock: new user is not already a member
	groupRepo.On("IsMember", groupID, newUserID).Return(false, nil)

	// Mock: add member
	groupRepo.On("AddMember", groupID, newUserID).Return(nil)

	err := svc.AddMember(groupID, newUserID, requestorID)
	require.NoError(t, err)

	groupRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestGroupService_AddMember_AlreadyMember(t *testing.T) {
	groupRepo := mocks.NewMockGroupRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewGroupService(groupRepo, userRepo)

	groupID := uuid.New()
	requestorID := uuid.New()
	existingUserID := uuid.New()

	group := &domain.Group{
		ID:        groupID,
		Name:      "Test Group",
		CreatedBy: requestorID,
		Members:   []uuid.UUID{requestorID, existingUserID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	existingUser := &domain.User{ID: existingUserID, Username: "existing"}

	// Mock: group exists
	groupRepo.On("GetByID", groupID).Return(group, nil)

	// Mock: requestor is member
	groupRepo.On("IsMember", groupID, requestorID).Return(true, nil)

	// Mock: user exists
	userRepo.On("GetByID", existingUserID).Return(existingUser, nil)

	// Mock: user is already a member
	groupRepo.On("IsMember", groupID, existingUserID).Return(true, nil)

	err := svc.AddMember(groupID, existingUserID, requestorID)
	require.NoError(t, err) // No error, just idempotent

	groupRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestGroupService_AddMember_Unauthorized(t *testing.T) {
	groupRepo := mocks.NewMockGroupRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewGroupService(groupRepo, userRepo)

	groupID := uuid.New()
	requestorID := uuid.New()
	newUserID := uuid.New()

	group := &domain.Group{
		ID:        groupID,
		Name:      "Test Group",
		CreatedBy: uuid.New(), // Different creator
		Members:   []uuid.UUID{uuid.New()},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Mock: group exists
	groupRepo.On("GetByID", groupID).Return(group, nil)

	// Mock: requestor is not a member
	groupRepo.On("IsMember", groupID, requestorID).Return(false, nil)

	err := svc.AddMember(groupID, newUserID, requestorID)
	assert.ErrorIs(t, err, domain.ErrUnauthorized)

	groupRepo.AssertExpectations(t)
}

func TestGroupService_RemoveMember_ByCreator(t *testing.T) {
	groupRepo := mocks.NewMockGroupRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewGroupService(groupRepo, userRepo)

	groupID := uuid.New()
	creatorID := uuid.New()
	memberID := uuid.New()

	group := &domain.Group{
		ID:        groupID,
		Name:      "Test Group",
		CreatedBy: creatorID,
		Members:   []uuid.UUID{creatorID, memberID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Mock: group exists
	groupRepo.On("GetByID", groupID).Return(group, nil)

	// Mock: remove member
	groupRepo.On("RemoveMember", groupID, memberID).Return(nil)

	err := svc.RemoveMember(groupID, memberID, creatorID)
	require.NoError(t, err)

	groupRepo.AssertExpectations(t)
}

func TestGroupService_RemoveMember_Self(t *testing.T) {
	groupRepo := mocks.NewMockGroupRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewGroupService(groupRepo, userRepo)

	groupID := uuid.New()
	creatorID := uuid.New()
	memberID := uuid.New()

	group := &domain.Group{
		ID:        groupID,
		Name:      "Test Group",
		CreatedBy: creatorID,
		Members:   []uuid.UUID{creatorID, memberID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Mock: group exists
	groupRepo.On("GetByID", groupID).Return(group, nil)

	// Mock: remove self
	groupRepo.On("RemoveMember", groupID, memberID).Return(nil)

	// Member removes themselves
	err := svc.RemoveMember(groupID, memberID, memberID)
	require.NoError(t, err)

	groupRepo.AssertExpectations(t)
}

func TestGroupService_RemoveMember_CannotRemoveCreator(t *testing.T) {
	groupRepo := mocks.NewMockGroupRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewGroupService(groupRepo, userRepo)

	groupID := uuid.New()
	creatorID := uuid.New()

	group := &domain.Group{
		ID:        groupID,
		Name:      "Test Group",
		CreatedBy: creatorID,
		Members:   []uuid.UUID{creatorID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Mock: group exists
	groupRepo.On("GetByID", groupID).Return(group, nil)

	// Try to remove creator
	err := svc.RemoveMember(groupID, creatorID, creatorID)
	assert.ErrorIs(t, err, domain.ErrForbidden)

	groupRepo.AssertExpectations(t)
}

func TestGroupService_RemoveMember_Unauthorized(t *testing.T) {
	groupRepo := mocks.NewMockGroupRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewGroupService(groupRepo, userRepo)

	groupID := uuid.New()
	creatorID := uuid.New()
	requestorID := uuid.New()
	memberID := uuid.New()

	group := &domain.Group{
		ID:        groupID,
		Name:      "Test Group",
		CreatedBy: creatorID,
		Members:   []uuid.UUID{creatorID, requestorID, memberID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Mock: group exists
	groupRepo.On("GetByID", groupID).Return(group, nil)

	// Non-creator tries to remove someone else
	err := svc.RemoveMember(groupID, memberID, requestorID)
	assert.ErrorIs(t, err, domain.ErrUnauthorized)

	groupRepo.AssertExpectations(t)
}

func TestGroupService_IsMember(t *testing.T) {
	groupRepo := mocks.NewMockGroupRepository(t)
	userRepo := mocks.NewMockUserRepository(t)

	svc := service.NewGroupService(groupRepo, userRepo)

	groupID := uuid.New()
	userID := uuid.New()

	// Mock: check membership
	groupRepo.On("IsMember", groupID, userID).Return(true, nil)

	isMember, err := svc.IsMember(groupID, userID)
	require.NoError(t, err)
	assert.True(t, isMember)

	groupRepo.AssertExpectations(t)
}

// =============================================================================
// Message Service Tests
// =============================================================================

func TestMessageService_SendConversationMessage_Success(t *testing.T) {
	mockMessageRepo := mocks.NewMockMessageRepository(t)
	mockConvRepo := mocks.NewMockConversationRepository(t)
	mockUserRepo := mocks.NewMockUserRepository(t)

	svc := service.NewMessageService(mockMessageRepo, mockConvRepo, mockUserRepo, nil)

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

	message, err := svc.SendConversationMessage(senderID, conversationID, content)

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
	mockMessageRepo := mocks.NewMockMessageRepository(t)
	mockConvRepo := mocks.NewMockConversationRepository(t)
	mockUserRepo := mocks.NewMockUserRepository(t)

	svc := service.NewMessageService(mockMessageRepo, mockConvRepo, mockUserRepo, nil)

	senderID := uuid.New()
	conversationID := uuid.New()
	content := "Hello!"

	// Mock: user exists
	mockUserRepo.On("GetByID", senderID).Return(&domain.User{ID: senderID}, nil)

	// Mock: conversation exists but user is NOT participant
	mockConvRepo.On("GetByID", conversationID).Return(&domain.Conversation{ID: conversationID}, nil)
	mockConvRepo.On("IsParticipant", conversationID, senderID).Return(false, nil)

	message, err := svc.SendConversationMessage(senderID, conversationID, content)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrUnauthorized, err)
	assert.Nil(t, message)

	mockUserRepo.AssertExpectations(t)
	mockConvRepo.AssertExpectations(t)
	mockMessageRepo.AssertNotCalled(t, "Create")
}

func TestMessageService_SendConversationMessage_EmptyContent(t *testing.T) {
	mockMessageRepo := mocks.NewMockMessageRepository(t)
	mockConvRepo := mocks.NewMockConversationRepository(t)
	mockUserRepo := mocks.NewMockUserRepository(t)

	svc := service.NewMessageService(mockMessageRepo, mockConvRepo, mockUserRepo, nil)

	senderID := uuid.New()
	conversationID := uuid.New()
	content := ""

	message, err := svc.SendConversationMessage(senderID, conversationID, content)

	assert.Error(t, err)
	assert.Nil(t, message)

	mockUserRepo.AssertNotCalled(t, "GetByID")
	mockConvRepo.AssertNotCalled(t, "GetByID")
	mockMessageRepo.AssertNotCalled(t, "Create")
}

func TestMessageService_SendGroupMessage_Success(t *testing.T) {
	mockMessageRepo := mocks.NewMockMessageRepository(t)
	mockGroupRepo := mocks.NewMockGroupRepository(t)
	mockUserRepo := mocks.NewMockUserRepository(t)

	svc := service.NewMessageService(mockMessageRepo, nil, mockUserRepo, mockGroupRepo)

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

	message, err := svc.SendGroupMessage(senderID, groupID, content)

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
	mockMessageRepo := mocks.NewMockMessageRepository(t)
	mockGroupRepo := mocks.NewMockGroupRepository(t)
	mockUserRepo := mocks.NewMockUserRepository(t)

	svc := service.NewMessageService(mockMessageRepo, nil, mockUserRepo, mockGroupRepo)

	senderID := uuid.New()
	groupID := uuid.New()
	content := "Hello group!"

	// Mock: user exists
	mockUserRepo.On("GetByID", senderID).Return(&domain.User{ID: senderID}, nil)

	// Mock: group exists but user is NOT member
	mockGroupRepo.On("GetByID", groupID).Return(&domain.Group{ID: groupID}, nil)
	mockGroupRepo.On("IsMember", groupID, senderID).Return(false, nil)

	message, err := svc.SendGroupMessage(senderID, groupID, content)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrUnauthorized, err)
	assert.Nil(t, message)

	mockUserRepo.AssertExpectations(t)
	mockGroupRepo.AssertExpectations(t)
	mockMessageRepo.AssertNotCalled(t, "Create")
}

func TestMessageService_GetConversationMessages_Success(t *testing.T) {
	mockMessageRepo := mocks.NewMockMessageRepository(t)
	mockConvRepo := mocks.NewMockConversationRepository(t)
	mockUserRepo := mocks.NewMockUserRepository(t)

	svc := service.NewMessageService(mockMessageRepo, mockConvRepo, mockUserRepo, nil)

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

	messages, err := svc.GetConversationMessages(userID, conversationID, nil, limit)

	assert.NoError(t, err)
	assert.NotNil(t, messages)
	assert.Len(t, messages, 2)

	mockUserRepo.AssertExpectations(t)
	mockConvRepo.AssertExpectations(t)
	mockMessageRepo.AssertExpectations(t)
}

func TestMessageService_GetConversationMessages_NotParticipant(t *testing.T) {
	mockMessageRepo := mocks.NewMockMessageRepository(t)
	mockConvRepo := mocks.NewMockConversationRepository(t)
	mockUserRepo := mocks.NewMockUserRepository(t)

	svc := service.NewMessageService(mockMessageRepo, mockConvRepo, mockUserRepo, nil)

	userID := uuid.New()
	conversationID := uuid.New()

	// Mock: user exists
	mockUserRepo.On("GetByID", userID).Return(&domain.User{ID: userID}, nil)

	// Mock: conversation exists but user is NOT participant
	mockConvRepo.On("GetByID", conversationID).Return(&domain.Conversation{ID: conversationID}, nil)
	mockConvRepo.On("IsParticipant", conversationID, userID).Return(false, nil)

	messages, err := svc.GetConversationMessages(userID, conversationID, nil, 20)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrUnauthorized, err)
	assert.Nil(t, messages)

	mockUserRepo.AssertExpectations(t)
	mockConvRepo.AssertExpectations(t)
	mockMessageRepo.AssertNotCalled(t, "ListByConversationID")
}

func TestMessageService_GetGroupMessages_Success(t *testing.T) {
	mockMessageRepo := mocks.NewMockMessageRepository(t)
	mockGroupRepo := mocks.NewMockGroupRepository(t)
	mockUserRepo := mocks.NewMockUserRepository(t)

	svc := service.NewMessageService(mockMessageRepo, nil, mockUserRepo, mockGroupRepo)

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

	messages, err := svc.GetGroupMessages(userID, groupID, nil, limit)

	assert.NoError(t, err)
	assert.NotNil(t, messages)
	assert.Len(t, messages, 2)

	mockUserRepo.AssertExpectations(t)
	mockGroupRepo.AssertExpectations(t)
	mockMessageRepo.AssertExpectations(t)
}

func TestMessageService_GetGroupMessages_NotMember(t *testing.T) {
	mockMessageRepo := mocks.NewMockMessageRepository(t)
	mockGroupRepo := mocks.NewMockGroupRepository(t)
	mockUserRepo := mocks.NewMockUserRepository(t)

	svc := service.NewMessageService(mockMessageRepo, nil, mockUserRepo, mockGroupRepo)

	userID := uuid.New()
	groupID := uuid.New()

	// Mock: user exists
	mockUserRepo.On("GetByID", userID).Return(&domain.User{ID: userID}, nil)

	// Mock: group exists but user is NOT member
	mockGroupRepo.On("GetByID", groupID).Return(&domain.Group{ID: groupID}, nil)
	mockGroupRepo.On("IsMember", groupID, userID).Return(false, nil)

	messages, err := svc.GetGroupMessages(userID, groupID, nil, 20)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrUnauthorized, err)
	assert.Nil(t, messages)

	mockUserRepo.AssertExpectations(t)
	mockGroupRepo.AssertExpectations(t)
	mockMessageRepo.AssertNotCalled(t, "ListByGroupID")
}
