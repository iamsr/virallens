package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ParseUUID parses a string into a UUID
func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// GetUserIDFromContext extracts the user ID from the Gin context
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, errors.New("user ID not found in context")
	}

	userIDStr, ok := userIDVal.(string)
	if !ok {
		return uuid.Nil, errors.New("user ID in context is not a string")
	}

	return uuid.Parse(userIDStr)
}
