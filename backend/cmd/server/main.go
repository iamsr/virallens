package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/yourusername/virallens/backend/internal/api"
	"github.com/yourusername/virallens/backend/internal/config"
	custommiddleware "github.com/yourusername/virallens/backend/internal/middleware"
	"github.com/yourusername/virallens/backend/internal/wire"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize application with Wire
	app, err := wire.InitializeApplication(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer app.DB.Close()

	// Run database migrations
	if err := config.RunMigrations(app.DB); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Start WebSocket hub
	go app.WebSocketHub.Run()
	log.Println("WebSocket hub started")

	// Create Echo server
	e := echo.New()
	e.HideBanner = true

	// Register custom validator
	e.Validator = custommiddleware.NewValidator()

	// Configure middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Configure server timeouts
	e.Server.ReadTimeout = cfg.Server.ReadTimeout
	e.Server.WriteTimeout = cfg.Server.WriteTimeout

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Setup API routes with Wire-injected dependencies
	api.SetupRoutes(
		e,
		app.AuthController,
		app.ConversationController,
		app.GroupController,
		app.JWTMiddleware,
		app.WebSocketHandler,
	)

	// Start server in a goroutine for graceful shutdown
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	go func() {
		log.Printf("Server starting on %s", serverAddr)
		if err := e.Start(serverAddr); err != nil {
			log.Printf("Server shutdown: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Gracefully shutdown the server
	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server shutdown complete")
}
