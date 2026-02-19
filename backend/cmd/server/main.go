package main

import (
	"database/sql"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/yourusername/virallens/backend/internal/config"
	"github.com/yourusername/virallens/backend/internal/repository"
	"github.com/yourusername/virallens/backend/internal/service"
)

// AppServices holds all application services
type AppServices struct {
	AuthService         service.AuthService
	ConversationService service.ConversationService
	GroupService        service.GroupService
	MessageService      *service.MessageService
	JWTService          service.JWTService
}

// InitializeApp creates all dependencies and returns the application services
func InitializeApp(db *sql.DB, jwtSecret string, accessTokenDuration, refreshTokenDuration time.Duration) *AppServices {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	conversationRepo := repository.NewConversationRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	// Initialize services
	jwtService := service.NewJWTService(jwtSecret, accessTokenDuration, refreshTokenDuration)
	authService := service.NewAuthService(userRepo, refreshTokenRepo, jwtService)
	conversationService := service.NewConversationService(conversationRepo, userRepo)
	groupService := service.NewGroupService(groupRepo, userRepo)
	messageService := service.NewMessageService(messageRepo, conversationRepo, userRepo, groupRepo)

	return &AppServices{
		AuthService:         authService,
		ConversationService: conversationService,
		GroupService:        groupService,
		MessageService:      messageService,
		JWTService:          jwtService,
	}
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Database connection
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := config.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// JWT configuration
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	accessTokenDuration, err := strconv.Atoi(os.Getenv("JWT_ACCESS_TOKEN_DURATION"))
	if err != nil || accessTokenDuration == 0 {
		accessTokenDuration = 15 // Default to 15 minutes
	}

	refreshTokenDuration, err := strconv.Atoi(os.Getenv("JWT_REFRESH_TOKEN_DURATION"))
	if err != nil || refreshTokenDuration == 0 {
		refreshTokenDuration = 10080 // Default to 7 days (7 * 24 * 60 minutes)
	}

	// Initialize services
	services := InitializeApp(
		db,
		jwtSecret,
		time.Duration(accessTokenDuration)*time.Minute,
		time.Duration(refreshTokenDuration)*time.Minute,
	)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// TODO: Register API routes here
	// Example: api.RegisterRoutes(e, services)
	_ = services // Suppress unused variable warning for now

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatal(err)
	}
}
