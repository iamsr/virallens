package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Define the core JWT interface here locally so we don't circularly depend on modules
// We'll pass the actual JWT service implementation during initialization
type JWTVerifier interface {
	ValidateAccessToken(token string) (userID string, err error)
}

// Authenticate returns a Gin middleware function that validates JWT tokens
func Authenticate(verifier JWTVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]

		userID, err := verifier.ValidateAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
