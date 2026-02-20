package dto

// CreateGroupRequest represents the request to create a group
type CreateGroupRequest struct {
	Name    string   `json:"name" validate:"required,min=3,max=100"`
	Members []string `json:"members" validate:"required,min=1,dive,uuid"`
}

// AddMemberRequest represents the request to add a member to a group
type AddMemberRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

// RemoveMemberRequest represents the request to remove a member from a group
type RemoveMemberRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}
