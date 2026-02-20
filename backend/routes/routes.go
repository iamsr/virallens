package routes

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/yourusername/virallens/backend/common/middlewares"
	"github.com/yourusername/virallens/backend/modules/auth"
	"github.com/yourusername/virallens/backend/modules/chat"
	"github.com/yourusername/virallens/backend/modules/user"
	"github.com/yourusername/virallens/backend/modules/websocket"
)

func SetupRouter(
	authCtrl *auth.Controller,
	userCtrl *user.Controller,
	convCtrl *chat.ConversationController,
	groupCtrl *chat.GroupController,
	wsHandler *websocket.Handler,
	jwtSvc auth.JWTService,
) *gin.Engine {
	r := gin.Default()

	// Initialize message rate limiter: 5 messages per 10 seconds
	msgRateLimiter := middlewares.NewRateLimiter(5, 10*time.Second)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000", "http://192.168.1.3:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", authCtrl.Register)
			authRoutes.POST("/login", authCtrl.Login)
			authRoutes.POST("/refresh", authCtrl.RefreshToken)
			authRoutes.POST("/logout", middlewares.Authenticate(jwtSvc), authCtrl.Logout)
		}

		userGroup := api.Group("/users")
		userGroup.Use(middlewares.Authenticate(jwtSvc))
		{
			userGroup.GET("", userCtrl.ListUsers)
		}

		convGroup := api.Group("/conversations")
		convGroup.Use(middlewares.Authenticate(jwtSvc))
		{
			convGroup.POST("", convCtrl.CreateOrGet)
			convGroup.GET("", convCtrl.List)
			convGroup.GET("/:id/messages", convCtrl.GetMessages)
			convGroup.POST("/:id/messages", msgRateLimiter.Middleware(), convCtrl.SendMessage)
		}

		grpGroup := api.Group("/groups")
		grpGroup.Use(middlewares.Authenticate(jwtSvc))
		{
			grpGroup.POST("", groupCtrl.Create)
			grpGroup.GET("", groupCtrl.List)
			grpGroup.GET("/:id", groupCtrl.Get)
			grpGroup.POST("/:id/members", groupCtrl.AddMember)
			grpGroup.DELETE("/:id/members", groupCtrl.RemoveMember)
			grpGroup.GET("/:id/messages", groupCtrl.GetMessages)
			grpGroup.POST("/:id/messages", msgRateLimiter.Middleware(), groupCtrl.SendMessage)
		}
	}

	r.GET("/ws", wsHandler.HandleWebSocket)

	return r
}
