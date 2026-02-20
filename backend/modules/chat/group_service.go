package chat

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/iamsr/virallens/backend/models"
	"github.com/iamsr/virallens/backend/modules/user"
)

type GroupService interface {
	Create(name string, createdByID uuid.UUID, memberIDs []uuid.UUID) (*models.Group, error)
	GetByID(groupID uuid.UUID) (*models.Group, error)
	ListUserGroups(userID uuid.UUID) ([]*models.Group, error)
	AddMember(adderID, groupID, userIDToAdd uuid.UUID) error
	RemoveMember(removerID, groupID, userIDToRemove uuid.UUID) error
}

type groupSvc struct {
	repo     GroupRepository
	userRepo user.Repository
}

func NewGroupService(repo GroupRepository, userRepo user.Repository) GroupService {
	return &groupSvc{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (s *groupSvc) Create(name string, createdByID uuid.UUID, memberIDs []uuid.UUID) (*models.Group, error) {
	if name == "" {
		return nil, errors.New("group name cannot be empty")
	}

	hasCreator := false
	for _, id := range memberIDs {
		if id == createdByID {
			hasCreator = true
			break
		}
	}
	if !hasCreator {
		memberIDs = append(memberIDs, createdByID)
	}

	group := &models.Group{
		ID:          uuid.New(),
		Name:        name,
		CreatedByID: createdByID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(group); err != nil {
		return nil, err
	}

	for _, memberID := range memberIDs {
		if err := s.repo.AddMember(group.ID, memberID); err != nil {
			return nil, err
		}
	}

	return group, nil
}

func (s *groupSvc) GetByID(groupID uuid.UUID) (*models.Group, error) {
	return s.repo.GetByID(groupID)
}

func (s *groupSvc) ListUserGroups(userID uuid.UUID) ([]*models.Group, error) {
	return s.repo.ListByUserID(userID)
}

func (s *groupSvc) AddMember(adderID, groupID, userIDToAdd uuid.UUID) error {
	isAdmin, err := s.isAdminOrCreator(groupID, adderID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return ErrUnauthorized
	}

	_, err = s.userRepo.GetByID(userIDToAdd)
	if err != nil {
		return errors.New("user not found")
	}

	isMember, err := s.repo.IsMember(groupID, userIDToAdd)
	if err != nil {
		return err
	}
	if isMember {
		return errors.New("user is already a member")
	}

	return s.repo.AddMember(groupID, userIDToAdd)
}

func (s *groupSvc) RemoveMember(removerID, groupID, userIDToRemove uuid.UUID) error {
	isAdmin, err := s.isAdminOrCreator(groupID, removerID)
	if err != nil {
		return err
	}

	if !isAdmin && removerID != userIDToRemove {
		return ErrUnauthorized
	}

	isMember, err := s.repo.IsMember(groupID, userIDToRemove)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("user is not a member")
	}

	return s.repo.RemoveMember(groupID, userIDToRemove)
}

func (s *groupSvc) isAdminOrCreator(groupID, userID uuid.UUID) (bool, error) {
	group, err := s.repo.GetByID(groupID)
	if err != nil {
		return false, err
	}
	return group.CreatedByID == userID, nil
}
