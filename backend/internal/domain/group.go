package domain

import (
	"time"

	"github.com/google/uuid"
)

type Group struct {
	ID        uuid.UUID   `json:"id"`
	Name      string      `json:"name"`
	CreatedBy uuid.UUID   `json:"created_by"`
	Members   []uuid.UUID `json:"members"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type GroupRepository interface {
	Create(group *Group) error
	GetByID(id uuid.UUID) (*Group, error)
	ListByUserID(userID uuid.UUID) ([]*Group, error)
	AddMember(groupID, userID uuid.UUID) error
	RemoveMember(groupID, userID uuid.UUID) error
	IsMember(groupID, userID uuid.UUID) (bool, error)
}
