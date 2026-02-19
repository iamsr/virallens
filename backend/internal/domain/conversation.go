package domain

import (
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID           uuid.UUID   `json:"id"`
	Participants []uuid.UUID `json:"participants"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type ConversationRepository interface {
	Create(conversation *Conversation) error
	GetByID(id uuid.UUID) (*Conversation, error)
	GetByParticipants(userID1, userID2 uuid.UUID) (*Conversation, error)
	ListByUserID(userID uuid.UUID) ([]*Conversation, error)
	AddParticipant(conversationID, userID uuid.UUID) error
}
