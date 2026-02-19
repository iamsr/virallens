package api

import (
	"github.com/labstack/echo/v4"
	"github.com/yourusername/virallens/backend/internal/middleware"
	"github.com/yourusername/virallens/backend/internal/websocket"
)

// SetupRoutes configures all API routes with Wire-injected dependencies
func SetupRoutes(
	e *echo.Echo,
	authController *AuthController,
	conversationController *ConversationController,
	groupController *GroupController,
	jwtMiddleware *middleware.JWTMiddleware,
	wsHandler *websocket.Handler,
) {
	// Public routes
	api := e.Group("/api")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.POST("/register", authController.Register)
	auth.POST("/login", authController.Login)
	auth.POST("/refresh", authController.RefreshToken)

	// Auth protected routes
	auth.POST("/logout", authController.Logout, jwtMiddleware.Authenticate)

	// Conversation routes
	conversations := api.Group("/conversations", jwtMiddleware.Authenticate)
	conversations.POST("", conversationController.CreateOrGet)
	conversations.GET("", conversationController.List)
	conversations.GET("/:id", conversationController.GetByID)
	conversations.GET("/:id/messages", conversationController.GetMessages)

	// Group routes
	groups := api.Group("/groups", jwtMiddleware.Authenticate)
	groups.POST("", groupController.Create)
	groups.GET("", groupController.List)
	groups.GET("/:id", groupController.GetByID)
	groups.POST("/:id/members", groupController.AddMember)
	groups.DELETE("/:id/members", groupController.RemoveMember)
	groups.GET("/:id/messages", groupController.GetMessages)

	// WebSocket route (requires authentication)
	e.GET("/ws", wsHandler.HandleWebSocket, jwtMiddleware.Authenticate)
}
