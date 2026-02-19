package api

import (
	"github.com/labstack/echo/v4"
	"github.com/yourusername/virallens/backend/internal/middleware"
	"github.com/yourusername/virallens/backend/internal/service"
	"github.com/yourusername/virallens/backend/internal/websocket"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(
	e *echo.Echo,
	authService service.AuthService,
	conversationService service.ConversationService,
	groupService service.GroupService,
	messageService *service.MessageService,
	jwtService service.JWTService,
	wsHandler *websocket.Handler,
) {
	// Create controllers
	authController := NewAuthController(authService)
	conversationController := NewConversationController(conversationService, messageService)
	groupController := NewGroupController(groupService, messageService)

	// Public routes
	api := e.Group("/api")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.POST("/register", authController.Register)
	auth.POST("/login", authController.Login)
	auth.POST("/refresh", authController.RefreshToken)

	// Protected routes (require JWT)
	jwtMiddleware := middleware.JWTMiddleware(jwtService)

	// Auth protected routes
	auth.POST("/logout", authController.Logout, jwtMiddleware)

	// Conversation routes
	conversations := api.Group("/conversations", jwtMiddleware)
	conversations.POST("", conversationController.CreateOrGet)
	conversations.GET("", conversationController.List)
	conversations.GET("/:id", conversationController.GetByID)
	conversations.GET("/:id/messages", conversationController.GetMessages)

	// Group routes
	groups := api.Group("/groups", jwtMiddleware)
	groups.POST("", groupController.Create)
	groups.GET("", groupController.List)
	groups.GET("/:id", groupController.GetByID)
	groups.POST("/:id/members", groupController.AddMember)
	groups.DELETE("/:id/members", groupController.RemoveMember)
	groups.GET("/:id/messages", groupController.GetMessages)

	// WebSocket route
	e.GET("/ws", wsHandler.HandleWebSocket)
}
