package wire

import (
	"database/sql"

	"github.com/yourusername/virallens/backend/internal/config"
	"github.com/yourusername/virallens/backend/internal/domain"
	"github.com/yourusername/virallens/backend/internal/middleware"
	"github.com/yourusername/virallens/backend/internal/service"
	"github.com/yourusername/virallens/backend/internal/websocket"
)

// ProvideDatabase creates database connection from config
func ProvideDatabase(cfg *config.Config) (*sql.DB, error) {
	return config.NewDatabase(&cfg.Database)
}

// ProvideJWTService creates JWT service from config
func ProvideJWTService(cfg *config.Config) service.JWTService {
	return service.NewJWTService(
		cfg.JWT.AccessSecret,
		cfg.JWT.AccessExpiration,
		cfg.JWT.RefreshExpiration,
	)
}

// ProvideJWTMiddleware creates JWT middleware
func ProvideJWTMiddleware(jwtService service.JWTService) *middleware.JWTMiddleware {
	return middleware.NewJWTMiddleware(jwtService)
}

// ProvideWebSocketHub creates WebSocket hub
func ProvideWebSocketHub() *websocket.Hub {
	return websocket.NewHub()
}

// ProvideMessageService creates message service (it returns a pointer, not interface)
func ProvideMessageService(
	messageRepo domain.MessageRepository,
	conversationRepo domain.ConversationRepository,
	userRepo domain.UserRepository,
	groupRepo domain.GroupRepository,
) *service.MessageService {
	return service.NewMessageService(messageRepo, conversationRepo, userRepo, groupRepo)
}
