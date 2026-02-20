package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yourusername/virallens/backend/internal/api/dto"
	"github.com/yourusername/virallens/backend/internal/service"
)

type AuthController struct {
	authService service.AuthService
}

func NewAuthController(authService service.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// Register handles user registration
func (ac *AuthController) Register(c echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	// Validate using Echo's validator
	if err := c.Validate(&req); err != nil {
		return err
	}

	resp, err := ac.authService.Register(&service.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"user":          resp.User,
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
	})
}

// Login handles user login
func (ac *AuthController) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	// Validate using Echo's validator
	if err := c.Validate(&req); err != nil {
		return err
	}

	resp, err := ac.authService.Login(&service.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user":          resp.User,
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
	})
}

// RefreshToken handles token refresh
func (ac *AuthController) RefreshToken(c echo.Context) error {
	var req dto.RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	// Validate using Echo's validator
	if err := c.Validate(&req); err != nil {
		return err
	}

	resp, err := ac.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"access_token": resp.AccessToken,
	})
}

// Logout handles user logout
func (ac *AuthController) Logout(c echo.Context) error {
	// Get user ID from context (set by JWT middleware)
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	uid, err := parseUUID(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
	}

	if err := ac.authService.Logout(uid); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "logged out successfully"})
}
