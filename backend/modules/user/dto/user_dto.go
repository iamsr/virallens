package dto

import (
	"time"

	"github.com/yourusername/virallens/backend/models"
)

type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func MapDomainUserToResponse(u *models.User) UserResponse {
	return UserResponse{
		ID:        u.ID.String(),
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}
}

func MapDomainUsersToResponse(users []*models.User) []UserResponse {
	response := make([]UserResponse, 0, len(users))
	for _, u := range users {
		response = append(response, MapDomainUserToResponse(u))
	}
	return response
}
