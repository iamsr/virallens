package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Group struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string         `gorm:"not null;size:100" json:"name"`
	CreatedByID uuid.UUID      `gorm:"type:uuid;not null" json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Creator User `gorm:"foreignKey:CreatedByID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
	Members []User `gorm:"many2many:group_members;" json:"-"`
}

type GroupMember struct {
	GroupID  uuid.UUID `gorm:"type:uuid;primaryKey;index" json:"group_id"`
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey;index" json:"user_id"`
	JoinedAt time.Time `gorm:"autoCreateTime" json:"joined_at"`
}
