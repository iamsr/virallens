package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/iamsr/virallens/backend/models"
)

type CreateOrGetRequest struct {
	OtherUserID uuid.UUID `json:"other_user_id" binding:"required"`
}

type GetMessagesQuery struct {
	Cursor *time.Time `form:"cursor"`
	Limit  int        `form:"limit"`
}

type CreateGroupRequest struct {
	Name    string      `json:"name" binding:"required,min=3,max=100"`
	Members []uuid.UUID `json:"members" binding:"required,min=1"`
}

type AddMemberRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

type RemoveMemberRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

type SendMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

// ConversationResponse mapped to models.Conversation
type ConversationResponse struct {
	ID           string    `json:"id"`
	Participants []string  `json:"participants"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func MapConversationToResponse(c *models.Conversation) ConversationResponse {
	return ConversationResponse{
		ID:           c.ID.String(),
		Participants: []string{c.Participant1.String(), c.Participant2.String()},
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}

// GroupResponse mapped to models.Group
type GroupResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Members     []string  `json:"members"`
	CreatedByID string    `json:"created_by_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func MapGroupToResponse(g *models.Group) GroupResponse {
	members := make([]string, 0, len(g.Members))
	for _, m := range g.Members {
		members = append(members, m.ID.String())
	}
	return GroupResponse{
		ID:          g.ID.String(),
		Name:        g.Name,
		Members:     members,
		CreatedByID: g.CreatedByID.String(),
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
}

func MapGroupsToResponse(groups []*models.Group) []GroupResponse {
	resp := make([]GroupResponse, 0, len(groups))
	for _, g := range groups {
		resp = append(resp, MapGroupToResponse(g))
	}
	return resp
}

// MessageResponse mapped to models.Message
type MessageResponse struct {
	ID             string     `json:"id"`
	SenderID       string     `json:"sender_id"`
	ConversationID *string    `json:"conversation_id,omitempty"`
	GroupID        *string    `json:"group_id,omitempty"`
	Content        string     `json:"content"`
	Type           string     `json:"type"`
	CreatedAt      time.Time  `json:"created_at"`
}

func MapMessageToResponse(m *models.Message) MessageResponse {
	resp := MessageResponse{
		ID:        m.ID.String(),
		SenderID:  m.SenderID.String(),
		Content:   m.Content,
		Type:      string(m.Type),
		CreatedAt: m.CreatedAt,
	}
	if m.ConversationID != nil {
		cid := m.ConversationID.String()
		resp.ConversationID = &cid
	}
	if m.GroupID != nil {
		gid := m.GroupID.String()
		resp.GroupID = &gid
	}
	return resp
}

func MapMessagesToResponse(messages []*models.Message) []MessageResponse {
	resp := make([]MessageResponse, 0, len(messages))
	for _, m := range messages {
		resp = append(resp, MapMessageToResponse(m))
	}
	return resp
}
