package user

import (
	"github.com/google/uuid"
	"github.com/yourusername/virallens/backend/models"
)

type Service interface {
	ListUsers(excludeUserID uuid.UUID) ([]*models.User, error)
}

type service struct {
	userRepo Repository
}

func NewService(userRepo Repository) Service {
	return &service{userRepo: userRepo}
}

func (s *service) ListUsers(excludeUserID uuid.UUID) ([]*models.User, error) {
	allUsers, err := s.userRepo.List()
	if err != nil {
		return nil, err
	}

	var filteredUsers []*models.User
	for _, u := range allUsers {
		if u.ID != excludeUserID {
			filteredUsers = append(filteredUsers, u)
		}
	}

	return filteredUsers, nil
}
