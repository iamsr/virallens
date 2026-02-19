package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/virallens/backend/internal/domain"
)

type GroupService interface {
	Create(name string, creatorID uuid.UUID, memberIDs []uuid.UUID) (*domain.Group, error)
	GetByID(id uuid.UUID, userID uuid.UUID) (*domain.Group, error)
	ListByUserID(userID uuid.UUID) ([]*domain.Group, error)
	AddMember(groupID, userID, requestorID uuid.UUID) error
	RemoveMember(groupID, userID, requestorID uuid.UUID) error
	IsMember(groupID, userID uuid.UUID) (bool, error)
}

type groupService struct {
	groupRepo domain.GroupRepository
	userRepo  domain.UserRepository
}

// NewGroupService creates a new group service
func NewGroupService(
	groupRepo domain.GroupRepository,
	userRepo domain.UserRepository,
) GroupService {
	return &groupService{
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

func (s *groupService) Create(name string, creatorID uuid.UUID, memberIDs []uuid.UUID) (*domain.Group, error) {
	// Validate creator exists
	_, err := s.userRepo.GetByID(creatorID)
	if err != nil {
		return nil, err
	}

	// Validate all members exist
	for _, memberID := range memberIDs {
		_, err := s.userRepo.GetByID(memberID)
		if err != nil {
			return nil, err
		}
	}

	// Ensure creator is in the member list
	creatorInList := false
	for _, memberID := range memberIDs {
		if memberID == creatorID {
			creatorInList = true
			break
		}
	}

	if !creatorInList {
		memberIDs = append(memberIDs, creatorID)
	}

	// Create group
	group := &domain.Group{
		ID:        uuid.New(),
		Name:      name,
		CreatedBy: creatorID,
		Members:   memberIDs,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.groupRepo.Create(group); err != nil {
		return nil, err
	}

	return group, nil
}

func (s *groupService) GetByID(id uuid.UUID, userID uuid.UUID) (*domain.Group, error) {
	group, err := s.groupRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Verify user is a member
	isMember, err := s.groupRepo.IsMember(id, userID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, domain.ErrNotGroupMember
	}

	return group, nil
}

func (s *groupService) ListByUserID(userID uuid.UUID) ([]*domain.Group, error) {
	// Verify user exists
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	return s.groupRepo.ListByUserID(userID)
}

func (s *groupService) AddMember(groupID, userID, requestorID uuid.UUID) error {
	// Get group
	group, err := s.groupRepo.GetByID(groupID)
	if err != nil {
		return err
	}

	// Verify requestor is the creator or a member
	isMember, err := s.groupRepo.IsMember(groupID, requestorID)
	if err != nil {
		return err
	}

	if !isMember && group.CreatedBy != requestorID {
		return domain.ErrUnauthorized
	}

	// Verify user to add exists
	_, err = s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// Check if user is already a member
	isAlreadyMember, err := s.groupRepo.IsMember(groupID, userID)
	if err != nil {
		return err
	}

	if isAlreadyMember {
		return nil // Already a member, no error
	}

	// Add member
	return s.groupRepo.AddMember(groupID, userID)
}

func (s *groupService) RemoveMember(groupID, userID, requestorID uuid.UUID) error {
	// Get group
	group, err := s.groupRepo.GetByID(groupID)
	if err != nil {
		return err
	}

	// Only the creator can remove members, or members can remove themselves
	if group.CreatedBy != requestorID && userID != requestorID {
		return domain.ErrUnauthorized
	}

	// Cannot remove the creator
	if userID == group.CreatedBy {
		return domain.ErrForbidden
	}

	// Remove member
	return s.groupRepo.RemoveMember(groupID, userID)
}

func (s *groupService) IsMember(groupID, userID uuid.UUID) (bool, error) {
	return s.groupRepo.IsMember(groupID, userID)
}
