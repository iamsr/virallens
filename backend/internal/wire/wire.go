//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/iamsr/virallens/backend/internal/config"
	"github.com/iamsr/virallens/backend/internal/db"
	"github.com/iamsr/virallens/backend/routes"
)

// InitializeServer sets up the Gin server with all dependencies injected.
func InitializeServer(cfg *config.Config) (*gin.Engine, error) {
	wire.Build(
		db.NewDatabase,

		UserSet,
		AuthSet,
		ChatSet,
		WebSocketSet,

		routes.SetupRouter,
	)
	return nil, nil
}
