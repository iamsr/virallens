package chat

import (
	"github.com/google/uuid"
	"github.com/yourusername/virallens/backend/models"
	"gorm.io/gorm"
)

type GroupRepository interface {
	Create(group *models.Group) error
	GetByID(id uuid.UUID) (*models.Group, error)
	ListByUserID(userID uuid.UUID) ([]*models.Group, error)
	AddMember(groupID, userID uuid.UUID) error
	RemoveMember(groupID, userID uuid.UUID) error
	IsMember(groupID, userID uuid.UUID) (bool, error)
}

type groupRepo struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) GroupRepository {
	return &groupRepo{db: db}
}

func (r *groupRepo) Create(group *models.Group) error {
	// GORM will automatically create the associations if they are populated
	return r.db.Create(group).Error
}

func (r *groupRepo) GetByID(id uuid.UUID) (*models.Group, error) {
	var group models.Group
	err := r.db.Preload("Members").First(&group, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *groupRepo) ListByUserID(userID uuid.UUID) ([]*models.Group, error) {
	var groups []*models.Group
	// Using Joins to find groups where user is a member
	err := r.db.Preload("Members").
		Joins("JOIN group_members ON group_members.group_id = groups.id").
		Where("group_members.user_id = ?", userID).
		Order("groups.updated_at desc").
		Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (r *groupRepo) AddMember(groupID, userID uuid.UUID) error {
	member := models.GroupMember{
		GroupID: groupID,
		UserID:  userID,
	}
	return r.db.Create(&member).Error
}

func (r *groupRepo) RemoveMember(groupID, userID uuid.UUID) error {
	return r.db.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&models.GroupMember{}).Error
}

func (r *groupRepo) IsMember(groupID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
