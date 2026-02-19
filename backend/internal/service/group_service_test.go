package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/virallens/backend/internal/domain"
)

// Mock GroupRepository
type MockGroupRepository struct {
	mock.Mock
}

func (m *MockGroupRepository) Create(group *domain.Group) error {
	args := m.Called(group)
	return args.Error(0)
}

func (m *MockGroupRepository) GetByID(id uuid.UUID) (*domain.Group, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Group), args.Error(1)
}

func (m *MockGroupRepository) ListByUserID(userID uuid.UUID) ([]*domain.Group, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Group), args.Error(1)
}

func (m *MockGroupRepository) AddMember(groupID, userID uuid.UUID) error {
	args := m.Called(groupID, userID)
	return args.Error(0)
}

func (m *MockGroupRepository) RemoveMember(groupID, userID uuid.UUID) error {
	args := m.Called(groupID, userID)
	return args.Error(0)
}

func (m *MockGroupRepository) IsMember(groupID, userID uuid.UUID) (bool, error) {
	args := m.Called(groupID, userID)
	return args.Bool(0), args.Error(1)
}

func TestGroupService_Create_Success(t *testing.T) {
	groupRepo := new(MockGroupRepository)
	userRepo := new(MockUserRepository)

	service := NewGroupService(groupRepo, userRepo)

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

	group, err := service.Create("Test Group", creatorID, []uuid.UUID{member1ID, member2ID})
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
	groupRepo := new(MockGroupRepository)
	userRepo := new(MockUserRepository)

	service := NewGroupService(groupRepo, userRepo)

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
	group, err := service.Create("Test Group", creatorID, []uuid.UUID{creatorID, member1ID})
	require.NoError(t, err)
	assert.Len(t, group.Members, 2) // No duplicate

	userRepo.AssertExpectations(t)
	groupRepo.AssertExpectations(t)
}

func TestGroupService_Create_MemberNotFound(t *testing.T) {
	groupRepo := new(MockGroupRepository)
	userRepo := new(MockUserRepository)

	service := NewGroupService(groupRepo, userRepo)

	creatorID := uuid.New()
	invalidMemberID := uuid.New()

	creator := &domain.User{ID: creatorID, Username: "creator"}

	// Mock: creator exists, member doesn't
	userRepo.On("GetByID", creatorID).Return(creator, nil)
	userRepo.On("GetByID", invalidMemberID).Return(nil, domain.ErrUserNotFound)

	_, err := service.Create("Test Group", creatorID, []uuid.UUID{invalidMemberID})
	assert.ErrorIs(t, err, domain.ErrUserNotFound)

	userRepo.AssertExpectations(t)
}

func TestGroupService_GetByID_Success(t *testing.T) {
	groupRepo := new(MockGroupRepository)
	userRepo := new(MockUserRepository)

	service := NewGroupService(groupRepo, userRepo)

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

	result, err := service.GetByID(groupID, userID)
	require.NoError(t, err)
	assert.Equal(t, groupID, result.ID)

	groupRepo.AssertExpectations(t)
}

func TestGroupService_GetByID_NotMember(t *testing.T) {
	groupRepo := new(MockGroupRepository)
	userRepo := new(MockUserRepository)

	service := NewGroupService(groupRepo, userRepo)

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

	_, err := service.GetByID(groupID, userID)
	assert.ErrorIs(t, err, domain.ErrNotGroupMember)

	groupRepo.AssertExpectations(t)
}

func TestGroupService_ListByUserID_Success(t *testing.T) {
	groupRepo := new(MockGroupRepository)
	userRepo := new(MockUserRepository)

	service := NewGroupService(groupRepo, userRepo)

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

	result, err := service.ListByUserID(userID)
	require.NoError(t, err)
	assert.Len(t, result, 2)

	userRepo.AssertExpectations(t)
	groupRepo.AssertExpectations(t)
}

func TestGroupService_AddMember_Success(t *testing.T) {
	groupRepo := new(MockGroupRepository)
	userRepo := new(MockUserRepository)

	service := NewGroupService(groupRepo, userRepo)

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

	err := service.AddMember(groupID, newUserID, requestorID)
	require.NoError(t, err)

	groupRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestGroupService_AddMember_AlreadyMember(t *testing.T) {
	groupRepo := new(MockGroupRepository)
	userRepo := new(MockUserRepository)

	service := NewGroupService(groupRepo, userRepo)

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

	err := service.AddMember(groupID, existingUserID, requestorID)
	require.NoError(t, err) // No error, just idempotent

	groupRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestGroupService_AddMember_Unauthorized(t *testing.T) {
	groupRepo := new(MockGroupRepository)
	userRepo := new(MockUserRepository)

	service := NewGroupService(groupRepo, userRepo)

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

	err := service.AddMember(groupID, newUserID, requestorID)
	assert.ErrorIs(t, err, domain.ErrUnauthorized)

	groupRepo.AssertExpectations(t)
}

func TestGroupService_RemoveMember_ByCreator(t *testing.T) {
	groupRepo := new(MockGroupRepository)
	userRepo := new(MockUserRepository)

	service := NewGroupService(groupRepo, userRepo)

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

	err := service.RemoveMember(groupID, memberID, creatorID)
	require.NoError(t, err)

	groupRepo.AssertExpectations(t)
}

func TestGroupService_RemoveMember_Self(t *testing.T) {
	groupRepo := new(MockGroupRepository)
	userRepo := new(MockUserRepository)

	service := NewGroupService(groupRepo, userRepo)

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
	err := service.RemoveMember(groupID, memberID, memberID)
	require.NoError(t, err)

	groupRepo.AssertExpectations(t)
}

func TestGroupService_RemoveMember_CannotRemoveCreator(t *testing.T) {
	groupRepo := new(MockGroupRepository)
	userRepo := new(MockUserRepository)

	service := NewGroupService(groupRepo, userRepo)

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
	err := service.RemoveMember(groupID, creatorID, creatorID)
	assert.ErrorIs(t, err, domain.ErrForbidden)

	groupRepo.AssertExpectations(t)
}

func TestGroupService_RemoveMember_Unauthorized(t *testing.T) {
	groupRepo := new(MockGroupRepository)
	userRepo := new(MockUserRepository)

	service := NewGroupService(groupRepo, userRepo)

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
	err := service.RemoveMember(groupID, memberID, requestorID)
	assert.ErrorIs(t, err, domain.ErrUnauthorized)

	groupRepo.AssertExpectations(t)
}

func TestGroupService_IsMember(t *testing.T) {
	groupRepo := new(MockGroupRepository)
	userRepo := new(MockUserRepository)

	service := NewGroupService(groupRepo, userRepo)

	groupID := uuid.New()
	userID := uuid.New()

	// Mock: check membership
	groupRepo.On("IsMember", groupID, userID).Return(true, nil)

	isMember, err := service.IsMember(groupID, userID)
	require.NoError(t, err)
	assert.True(t, isMember)

	groupRepo.AssertExpectations(t)
}
