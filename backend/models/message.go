package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MessageType string

const (
	MessageTypeConversation MessageType = "conversation"
	MessageTypeGroup        MessageType = "group"
)

type Message struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	SenderID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"sender_id"`
	ConversationID *uuid.UUID     `gorm:"type:uuid;index" json:"conversation_id,omitempty"`
	GroupID        *uuid.UUID     `gorm:"type:uuid;index" json:"group_id,omitempty"`
	Content        string         `gorm:"type:text;not null" json:"content"`
	Type           MessageType    `gorm:"type:varchar(20);not null" json:"type"`
	CreatedAt      time.Time      `gorm:"index" json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Sender       User          `gorm:"foreignKey:SenderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Conversation *Conversation `gorm:"foreignKey:ConversationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Group        *Group        `gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}
