package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/yourusername/virallens/backend/internal/service"
)

// JWTMiddleware wraps JWT service for middleware functionality
type JWTMiddleware struct {
	jwtService service.JWTService
}

// NewJWTMiddleware creates a new JWT middleware
func NewJWTMiddleware(jwtService service.JWTService) *JWTMiddleware {
	return &JWTMiddleware{
		jwtService: jwtService,
	}
}

// Authenticate returns an Echo middleware function that validates JWT tokens
func (m *JWTMiddleware) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get token from Authorization header
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid authorization header format"})
		}

		token := parts[1]

		// Validate token
		claims, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
		}

		// Set user ID in context
		c.Set("user_id", claims.UserID.String())

		return next(c)
	}
}
