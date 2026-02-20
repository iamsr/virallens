package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Conversation struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Participant1 uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_conversation_participants" json:"participant_1"`
	Participant2 uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_conversation_participants" json:"participant_2"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	User1 User `gorm:"foreignKey:Participant1;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	User2 User `gorm:"foreignKey:Participant2;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}
