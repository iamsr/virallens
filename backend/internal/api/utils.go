package api

import (
	"errors"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// parseUUID parses a string into a UUID
func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// getUserIDFromContext extracts the user ID from the Echo context
func getUserIDFromContext(c echo.Context) (uuid.UUID, error) {
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return uuid.Nil, errors.New("user ID not found in context")
	}
	return uuid.Parse(userID)
}
