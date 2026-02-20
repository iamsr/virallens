package dto

// CreateOrGetRequest represents the request to create or get a conversation
type CreateOrGetRequest struct {
	OtherUserID string `json:"other_user_id" validate:"required,uuid"`
}

// GetMessagesQuery represents query parameters for getting messages
type GetMessagesQuery struct {
	Cursor string `query:"cursor"`
	Limit  int    `query:"limit" validate:"omitempty,min=1,max=100"`
}
