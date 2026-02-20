package chat

import (
	"time"

	"github.com/google/uuid"
	"github.com/iamsr/virallens/backend/models"
	"gorm.io/gorm"
)

type MessageRepository interface {
	Create(message *models.Message) error
	GetByID(id uuid.UUID) (*models.Message, error)
	ListByConversationID(conversationID uuid.UUID, cursor *time.Time, limit int) ([]*models.Message, error)
	ListByGroupID(groupID uuid.UUID, cursor *time.Time, limit int) ([]*models.Message, error)
}

type messageRepo struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepo{db: db}
}

func (r *messageRepo) Create(message *models.Message) error {
	// Start a transaction to create the message and update the parent's updated_at
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(message).Error; err != nil {
			return err
		}

		// Update parent's updated_at timestamp
		if message.ConversationID != nil {
			if err := tx.Model(&models.Conversation{}).Where("id = ?", *message.ConversationID).UpdateColumn("updated_at", message.CreatedAt).Error; err != nil {
				return err
			}
		} else if message.GroupID != nil {
			if err := tx.Model(&models.Group{}).Where("id = ?", *message.GroupID).UpdateColumn("updated_at", message.CreatedAt).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *messageRepo) GetByID(id uuid.UUID) (*models.Message, error) {
	var msg models.Message
	err := r.db.First(&msg, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *messageRepo) ListByConversationID(conversationID uuid.UUID, cursor *time.Time, limit int) ([]*models.Message, error) {
	var msgs []*models.Message
	query := r.db.Where("conversation_id = ?", conversationID).Order("created_at desc").Limit(limit)

	if cursor != nil {
		query = query.Where("created_at < ?", *cursor)
	}

	err := query.Find(&msgs).Error
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

func (r *messageRepo) ListByGroupID(groupID uuid.UUID, cursor *time.Time, limit int) ([]*models.Message, error) {
	var msgs []*models.Message
	query := r.db.Where("group_id = ?", groupID).Order("created_at desc").Limit(limit)

	if cursor != nil {
		query = query.Where("created_at < ?", *cursor)
	}

	err := query.Find(&msgs).Error
	if err != nil {
		return nil, err
	}
	return msgs, nil
}
