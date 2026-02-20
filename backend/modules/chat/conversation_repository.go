package chat

import (
	"github.com/google/uuid"
	"github.com/iamsr/virallens/backend/models"
	"gorm.io/gorm"
)

type ConversationRepository interface {
	Create(conversation *models.Conversation) error
	GetByID(id uuid.UUID) (*models.Conversation, error)
	GetByParticipants(user1ID, user2ID uuid.UUID) (*models.Conversation, error)
	ListByUserID(userID uuid.UUID) ([]*models.Conversation, error)
	IsParticipant(conversationID, userID uuid.UUID) (bool, error)
}

type conversationRepo struct {
	db *gorm.DB
}

func NewConversationRepository(db *gorm.DB) ConversationRepository {
	return &conversationRepo{db: db}
}

func (r *conversationRepo) Create(conversation *models.Conversation) error {
	return r.db.Create(conversation).Error
}

func (r *conversationRepo) GetByID(id uuid.UUID) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.First(&conv, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *conversationRepo) GetByParticipants(user1ID, user2ID uuid.UUID) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.Where(
		"(participant1 = ? AND participant2 = ?) OR (participant1 = ? AND participant2 = ?)",
		user1ID, user2ID, user2ID, user1ID,
	).First(&conv).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Return nil, nil if not found
		}
		return nil, err
	}
	return &conv, nil
}

func (r *conversationRepo) ListByUserID(userID uuid.UUID) ([]*models.Conversation, error) {
	var convs []*models.Conversation
	err := r.db.Where("participant1 = ? OR participant2 = ?", userID, userID).
		Order("updated_at desc").
		Find(&convs).Error
	if err != nil {
		return nil, err
	}
	return convs, nil
}

func (r *conversationRepo) IsParticipant(conversationID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Conversation{}).
		Where("id = ? AND (participant1 = ? OR participant2 = ?)", conversationID, userID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
