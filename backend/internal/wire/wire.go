//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/yourusername/virallens/backend/internal/api"
	"github.com/yourusername/virallens/backend/internal/config"
	"github.com/yourusername/virallens/backend/internal/middleware"
	"github.com/yourusername/virallens/backend/internal/websocket"
)

// Application holds all initialized components
type Application struct {
	Config                 *config.Config
	AuthController         *api.AuthController
	ConversationController *api.ConversationController
	GroupController        *api.GroupController
	JWTMiddleware          *middleware.JWTMiddleware
	WebSocketHub           *websocket.Hub
	WebSocketHandler       *websocket.Handler
}

// InitializeApplication wires up all dependencies
func InitializeApplication(cfg *config.Config) (*Application, error) {
	wire.Build(
		ProvideDatabase,
		RepositorySet,
		ServiceSet,
		ControllerSet,
		MiddlewareSet,
		WebSocketSet,
		wire.Struct(new(Application), "*"),
	)
	return nil, nil
}
