package domain

import (
	"time"

	"github.com/google/uuid"
)

type MessageType string

const (
	MessageTypeConversation MessageType = "conversation"
	MessageTypeGroup        MessageType = "group"
)

type Message struct {
	ID             uuid.UUID   `json:"id"`
	SenderID       uuid.UUID   `json:"sender_id"`
	ConversationID *uuid.UUID  `json:"conversation_id,omitempty"`
	GroupID        *uuid.UUID  `json:"group_id,omitempty"`
	Content        string      `json:"content"`
	Type           MessageType `json:"type"`
	CreatedAt      time.Time   `json:"created_at"`
}

type MessageRepository interface {
	Create(message *Message) error
	GetByID(id uuid.UUID) (*Message, error)
	ListByConversationID(conversationID uuid.UUID, cursor *time.Time, limit int) ([]*Message, error)
	ListByGroupID(groupID uuid.UUID, cursor *time.Time, limit int) ([]*Message, error)
}
