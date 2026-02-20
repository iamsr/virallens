package wire

import (
	"github.com/google/wire"
	"github.com/yourusername/virallens/backend/internal/config"

	"github.com/yourusername/virallens/backend/modules/auth"
	"github.com/yourusername/virallens/backend/modules/chat"
	"github.com/yourusername/virallens/backend/modules/user"
	"github.com/yourusername/virallens/backend/modules/websocket"
)

// ProvideJWTService provides a configured JWT service
func ProvideJWTService(cfg *config.Config) auth.JWTService {
	// Use config struct fields
	return auth.NewJWTService(cfg.JWT.AccessSecret, cfg.JWT.AccessExpiration, cfg.JWT.RefreshExpiration)
}

// AuthSet provides auth dependencies
var AuthSet = wire.NewSet(
	ProvideJWTService,
	auth.NewRefreshTokenRepository,
	auth.NewService,
	auth.NewController,
)

// UserSet provides user dependencies
var UserSet = wire.NewSet(
	user.NewRepository,
	user.NewService,
	user.NewController,
)

// ChatSet provides chat dependencies
var ChatSet = wire.NewSet(
	chat.NewConversationRepository,
	chat.NewGroupRepository,
	chat.NewMessageRepository,
	chat.NewConversationService,
	chat.NewGroupService,
	chat.NewMessageService,
	chat.NewConversationController,
	chat.NewGroupController,
)

// WebSocketSet provides websocket dependencies
var WebSocketSet = wire.NewSet(
	websocket.NewHub,
	websocket.NewHandler,
)
