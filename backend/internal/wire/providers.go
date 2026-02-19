//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/yourusername/virallens/backend/internal/api"
	"github.com/yourusername/virallens/backend/internal/repository"
	"github.com/yourusername/virallens/backend/internal/service"
	"github.com/yourusername/virallens/backend/internal/websocket"
)

// Repository providers
var RepositorySet = wire.NewSet(
	repository.NewUserRepository,
	repository.NewConversationRepository,
	repository.NewGroupRepository,
	repository.NewMessageRepository,
	repository.NewRefreshTokenRepository,
)

// Service providers
var ServiceSet = wire.NewSet(
	service.NewAuthService,
	service.NewConversationService,
	service.NewGroupService,
	ProvideMessageService,
	ProvideJWTService,
)

// Controller providers
var ControllerSet = wire.NewSet(
	api.NewAuthController,
	api.NewConversationController,
	api.NewGroupController,
)

// Middleware providers
var MiddlewareSet = wire.NewSet(
	ProvideJWTMiddleware,
)

// WebSocket providers
var WebSocketSet = wire.NewSet(
	ProvideWebSocketHub,
	websocket.NewHandler,
)
