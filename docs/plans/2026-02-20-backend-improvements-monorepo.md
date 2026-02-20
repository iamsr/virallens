# Backend Improvements & Monorepo Setup Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Transform the backend into a production-ready service with Wire DI, improved test structure, enhanced configuration, and set up a monorepo structure for easy development and deployment.

**Architecture:** Clean Architecture with Wire for compile-time dependency injection, centralized test infrastructure in `test/` directory, environment-based configuration with validation, and monorepo orchestration using npm workspaces and Makefiles.

**Tech Stack:** Go 1.25+, Google Wire, testify, Mockery, golang-migrate, Viper, zerolog, npm workspaces, Docker Compose

---

## Phase 1: Wire Dependency Injection Setup

### Task 1: Install Wire and Mockery Dependencies

**Files:**
- Modify: `backend/go.mod`
- Create: `backend/.mockery.yaml`

**Step 1: Install Wire and related dependencies**

Run:
```bash
cd backend
go get github.com/google/wire/cmd/wire
go get github.com/vektra/mockery/v2
go install github.com/google/wire/cmd/wire@latest
go install github.com/vektra/mockery/v2@latest
```

Expected: Dependencies installed successfully

**Step 2: Create Mockery configuration**

Create `backend/.mockery.yaml`:
```yaml
with-expecter: true
dir: "test/mocks"
outpkg: mocks
filename: "{{.InterfaceName}}_mock.go"
all: true
packages:
  github.com/yourusername/virallens/backend/internal/domain:
    interfaces:
      UserRepository:
      ConversationRepository:
      GroupRepository:
      MessageRepository:
      RefreshTokenRepository:
      AuthService:
      ConversationService:
      GroupService:
      MessageService:
      JWTService:
```

**Step 3: Verify dependencies**

Run: `cd backend && go mod tidy && go mod verify`
Expected: All dependencies verified

**Step 4: Commit**

```bash
git add backend/go.mod backend/go.sum backend/.mockery.yaml
git commit -m "feat: add Wire and Mockery dependencies for DI and testing"
```

---

### Task 2: Create Enhanced Config Package

**Files:**
- Create: `backend/internal/config/config.go`
- Create: `backend/internal/config/validator.go`
- Modify: `backend/internal/config/database.go`

**Step 1: Create typed Config struct**

Create `backend/internal/config/config.go`:
```go
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	App      AppConfig
}

type ServerConfig struct {
	Port            int
	Host            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type JWTConfig struct {
	AccessSecret      string
	RefreshSecret     string
	AccessExpiration  time.Duration
	RefreshExpiration time.Duration
}

type AppConfig struct {
	Environment string // development, production, test
	LogLevel    string // debug, info, warn, error
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnvAsInt("SERVER_PORT", 8080),
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:     getEnvAsDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getEnvAsDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			ShutdownTimeout: getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			DBName:          getEnv("DB_NAME", "virallens"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		JWT: JWTConfig{
			AccessSecret:      getEnv("JWT_ACCESS_SECRET", ""),
			RefreshSecret:     getEnv("JWT_REFRESH_SECRET", ""),
			AccessExpiration:  getEnvAsDuration("JWT_ACCESS_EXPIRATION", 15*time.Minute),
			RefreshExpiration: getEnvAsDuration("JWT_REFRESH_EXPIRATION", 7*24*time.Hour),
		},
		App: AppConfig{
			Environment: getEnv("APP_ENV", "development"),
			LogLevel:    getEnv("LOG_LEVEL", "info"),
		},
	}

	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// ConnectionString returns PostgreSQL connection string
func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
```

**Step 2: Create config validator**

Create `backend/internal/config/validator.go`:
```go
package config

import "errors"

// Validate checks if configuration is valid
func Validate(cfg *Config) error {
	if err := validateServer(&cfg.Server); err != nil {
		return err
	}
	if err := validateDatabase(&cfg.Database); err != nil {
		return err
	}
	if err := validateJWT(&cfg.JWT); err != nil {
		return err
	}
	if err := validateApp(&cfg.App); err != nil {
		return err
	}
	return nil
}

func validateServer(cfg *ServerConfig) error {
	if cfg.Port < 1 || cfg.Port > 65535 {
		return errors.New("server port must be between 1 and 65535")
	}
	if cfg.Host == "" {
		return errors.New("server host cannot be empty")
	}
	if cfg.ReadTimeout <= 0 {
		return errors.New("server read timeout must be positive")
	}
	if cfg.WriteTimeout <= 0 {
		return errors.New("server write timeout must be positive")
	}
	return nil
}

func validateDatabase(cfg *DatabaseConfig) error {
	if cfg.Host == "" {
		return errors.New("database host cannot be empty")
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return errors.New("database port must be between 1 and 65535")
	}
	if cfg.User == "" {
		return errors.New("database user cannot be empty")
	}
	if cfg.DBName == "" {
		return errors.New("database name cannot be empty")
	}
	if cfg.MaxOpenConns < 1 {
		return errors.New("database max open connections must be at least 1")
	}
	if cfg.MaxIdleConns < 0 {
		return errors.New("database max idle connections cannot be negative")
	}
	return nil
}

func validateJWT(cfg *JWTConfig) error {
	if cfg.AccessSecret == "" {
		return errors.New("JWT access secret cannot be empty")
	}
	if cfg.RefreshSecret == "" {
		return errors.New("JWT refresh secret cannot be empty")
	}
	if cfg.AccessExpiration <= 0 {
		return errors.New("JWT access expiration must be positive")
	}
	if cfg.RefreshExpiration <= 0 {
		return errors.New("JWT refresh expiration must be positive")
	}
	return nil
}

func validateApp(cfg *AppConfig) error {
	validEnvs := map[string]bool{
		"development": true,
		"production":  true,
		"test":        true,
	}
	if !validEnvs[cfg.Environment] {
		return errors.New("app environment must be development, production, or test")
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[cfg.LogLevel] {
		return errors.New("log level must be debug, info, warn, or error")
	}
	return nil
}
```

**Step 3: Update database.go to use Config**

Modify `backend/internal/config/database.go`:
```go
package config

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// NewDatabase creates a new database connection using Config
func NewDatabase(cfg *DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
```

Add import: `"context"`

**Step 4: Run tests to verify compilation**

Run: `cd backend && go build ./...`
Expected: Successful compilation

**Step 5: Commit**

```bash
git add backend/internal/config/
git commit -m "feat: add typed Config with env-based loading and validation"
```

---

### Task 3: Refactor Repository Structs to Public

**Files:**
- Modify: `backend/internal/repository/user_repository.go`
- Modify: `backend/internal/repository/conversation_repository.go`
- Modify: `backend/internal/repository/group_repository.go`
- Modify: `backend/internal/repository/message_repository.go`
- Modify: `backend/internal/repository/refresh_token_repository.go`

**Step 1: Refactor UserRepository**

In `backend/internal/repository/user_repository.go`:

Change:
```go
type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) domain.UserRepository {
	return &userRepository{db: db}
}
```

To:
```go
// UserRepositoryImpl implements domain.UserRepository
type UserRepositoryImpl struct {
	db *sql.DB
}

// Compile-time interface verification
var _ domain.UserRepository = (*UserRepositoryImpl)(nil)

// NewUserRepository creates a new UserRepositoryImpl
func NewUserRepository(db *sql.DB) domain.UserRepository {
	return &UserRepositoryImpl{db: db}
}
```

Update all method receivers from `(r *userRepository)` to `(r *UserRepositoryImpl)`

**Step 2: Refactor ConversationRepository**

In `backend/internal/repository/conversation_repository.go`:

Change:
```go
type conversationRepository struct {
	db *sql.DB
}

func NewConversationRepository(db *sql.DB) domain.ConversationRepository {
	return &conversationRepository{db: db}
}
```

To:
```go
// ConversationRepositoryImpl implements domain.ConversationRepository
type ConversationRepositoryImpl struct {
	db *sql.DB
}

// Compile-time interface verification
var _ domain.ConversationRepository = (*ConversationRepositoryImpl)(nil)

// NewConversationRepository creates a new ConversationRepositoryImpl
func NewConversationRepository(db *sql.DB) domain.ConversationRepository {
	return &ConversationRepositoryImpl{db: db}
}
```

Update all method receivers from `(r *conversationRepository)` to `(r *ConversationRepositoryImpl)`

**Step 3: Refactor GroupRepository**

In `backend/internal/repository/group_repository.go`:

Change:
```go
type groupRepository struct {
	db *sql.DB
}

func NewGroupRepository(db *sql.DB) domain.GroupRepository {
	return &groupRepository{db: db}
}
```

To:
```go
// GroupRepositoryImpl implements domain.GroupRepository
type GroupRepositoryImpl struct {
	db *sql.DB
}

// Compile-time interface verification
var _ domain.GroupRepository = (*GroupRepositoryImpl)(nil)

// NewGroupRepository creates a new GroupRepositoryImpl
func NewGroupRepository(db *sql.DB) domain.GroupRepository {
	return &GroupRepositoryImpl{db: db}
}
```

Update all method receivers from `(r *groupRepository)` to `(r *GroupRepositoryImpl)`

**Step 4: Refactor MessageRepository**

In `backend/internal/repository/message_repository.go`:

Change:
```go
type messageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) domain.MessageRepository {
	return &messageRepository{db: db}
}
```

To:
```go
// MessageRepositoryImpl implements domain.MessageRepository
type MessageRepositoryImpl struct {
	db *sql.DB
}

// Compile-time interface verification
var _ domain.MessageRepository = (*MessageRepositoryImpl)(nil)

// NewMessageRepository creates a new MessageRepositoryImpl
func NewMessageRepository(db *sql.DB) domain.MessageRepository {
	return &MessageRepositoryImpl{db: db}
}
```

Update all method receivers from `(r *messageRepository)` to `(r *MessageRepositoryImpl)`

**Step 5: Refactor RefreshTokenRepository**

In `backend/internal/repository/refresh_token_repository.go`:

Change:
```go
type refreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) domain.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}
```

To:
```go
// RefreshTokenRepositoryImpl implements domain.RefreshTokenRepository
type RefreshTokenRepositoryImpl struct {
	db *sql.DB
}

// Compile-time interface verification
var _ domain.RefreshTokenRepository = (*RefreshTokenRepositoryImpl)(nil)

// NewRefreshTokenRepository creates a new RefreshTokenRepositoryImpl
func NewRefreshTokenRepository(db *sql.DB) domain.RefreshTokenRepository {
	return &RefreshTokenRepositoryImpl{db: db}
}
```

Update all method receivers from `(r *refreshTokenRepository)` to `(r *RefreshTokenRepositoryImpl)`

**Step 6: Verify compilation**

Run: `cd backend && go build ./...`
Expected: Successful compilation

**Step 7: Commit**

```bash
git add backend/internal/repository/
git commit -m "refactor: make repository structs public with Impl suffix for Wire"
```

---

### Task 4: Create Wire Configuration

**Files:**
- Create: `backend/internal/wire/wire.go`
- Create: `backend/internal/wire/providers.go`

**Step 1: Create Wire providers**

Create `backend/internal/wire/providers.go`:
```go
//go:build wireinject
// +build wireinject

package wire

import (
	"database/sql"

	"github.com/yourusername/virallens/backend/internal/api"
	"github.com/yourusername/virallens/backend/internal/config"
	"github.com/yourusername/virallens/backend/internal/domain"
	"github.com/yourusername/virallens/backend/internal/middleware"
	"github.com/yourusername/virallens/backend/internal/repository"
	"github.com/yourusername/virallens/backend/internal/service"
	"github.com/yourusername/virallens/backend/internal/websocket"
	"github.com/google/wire"
)

// ProvideDatabase creates database connection from config
func ProvideDatabase(cfg *config.Config) (*sql.DB, error) {
	return config.NewDatabase(&cfg.Database)
}

// ProvideJWTService creates JWT service from config
func ProvideJWTService(cfg *config.Config) domain.JWTService {
	return service.NewJWTService(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessExpiration,
		cfg.JWT.RefreshExpiration,
	)
}

// ProvideJWTMiddleware creates JWT middleware
func ProvideJWTMiddleware(jwtService domain.JWTService) *middleware.JWTMiddleware {
	return middleware.NewJWTMiddleware(jwtService)
}

// ProvideWebSocketHub creates WebSocket hub
func ProvideWebSocketHub() *websocket.Hub {
	return websocket.NewHub()
}

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
	service.NewMessageService,
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
```

**Step 2: Create Wire injection configuration**

Create `backend/internal/wire/wire.go`:
```go
//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/yourusername/virallens/backend/internal/api"
	"github.com/yourusername/virallens/backend/internal/config"
	"github.com/yourusername/virallens/backend/internal/middleware"
	"github.com/yourusername/virallens/backend/internal/websocket"
	"github.com/google/wire"
)

// Application holds all initialized components
type Application struct {
	Config              *config.Config
	AuthController      *api.AuthController
	ConversationController *api.ConversationController
	GroupController     *api.GroupController
	JWTMiddleware       *middleware.JWTMiddleware
	WebSocketHub        *websocket.Hub
	WebSocketHandler    *websocket.Handler
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
```

**Step 3: Generate Wire code**

Run:
```bash
cd backend/internal/wire
wire
```

Expected: Output like "wire: github.com/yourusername/virallens/backend/internal/wire: wrote wire_gen.go"

**Step 4: Verify generated code exists**

Run: `ls backend/internal/wire/wire_gen.go`
Expected: File exists

**Step 5: Verify compilation**

Run: `cd backend && go build ./...`
Expected: Successful compilation

**Step 6: Commit**

```bash
git add backend/internal/wire/
git commit -m "feat: add Wire dependency injection configuration and generated code"
```

---

## Phase 2: Test Infrastructure Setup

### Task 5: Create Test Utilities

**Files:**
- Create: `backend/test/testutil/db.go`
- Create: `backend/test/testutil/factory.go`
- Create: `backend/test/testutil/assert.go`

**Step 1: Create database test utilities**

Create `backend/test/testutil/db.go`:
```go
package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// SetupTestDB creates a test database connection
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Fatal("TEST_DATABASE_URL environment variable not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping test database: %v", err)
	}

	return db
}

// CleanupTestDB cleans all tables in the test database
func CleanupTestDB(t *testing.T, db *sql.DB) {
	t.Helper()

	tables := []string{
		"messages",
		"group_members",
		"groups",
		"conversation_participants",
		"conversations",
		"refresh_tokens",
		"users",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Logf("warning: failed to truncate table %s: %v", table, err)
		}
	}
}

// TeardownTestDB closes the database connection
func TeardownTestDB(t *testing.T, db *sql.DB) {
	t.Helper()
	if err := db.Close(); err != nil {
		t.Logf("warning: failed to close database: %v", err)
	}
}
```

**Step 2: Create test factories**

Create `backend/test/testutil/factory.go`:
```go
package testutil

import (
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/virallens/backend/internal/domain"
)

// CreateTestUser creates a test user with default values
func CreateTestUser() *domain.User {
	return &domain.User{
		ID:          uuid.New(),
		Username:    "testuser_" + uuid.New().String()[:8],
		Email:       "test_" + uuid.New().String()[:8] + "@example.com",
		Password:    "$2a$10$abcdefghijklmnopqrstuv", // bcrypt hash
		DisplayName: "Test User",
		Bio:         "Test bio",
		AvatarURL:   "https://example.com/avatar.jpg",
		IsOnline:    false,
		LastSeenAt:  time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// CreateTestConversation creates a test conversation
func CreateTestConversation(user1ID, user2ID uuid.UUID) *domain.Conversation {
	return &domain.Conversation{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateTestGroup creates a test group
func CreateTestGroup(creatorID uuid.UUID) *domain.Group {
	return &domain.Group{
		ID:          uuid.New(),
		Name:        "Test Group " + uuid.New().String()[:8],
		Description: "Test group description",
		AvatarURL:   "https://example.com/group.jpg",
		CreatedBy:   creatorID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// CreateTestMessage creates a test message
func CreateTestMessage(senderID, conversationID uuid.UUID) *domain.Message {
	return &domain.Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        "Test message content",
		MessageType:    domain.MessageTypeText,
		IsRead:         false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}
```

**Step 3: Create custom assertions**

Create `backend/test/testutil/assert.go`:
```go
package testutil

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// AssertValidUUID checks if a UUID is valid and not nil
func AssertValidUUID(t *testing.T, id uuid.UUID, msgAndArgs ...interface{}) {
	t.Helper()
	assert.NotEqual(t, uuid.Nil, id, msgAndArgs...)
}

// AssertTimeRecent checks if a time is within the last minute
func AssertTimeRecent(t *testing.T, timestamp time.Time, msgAndArgs ...interface{}) {
	t.Helper()
	now := time.Now()
	diff := now.Sub(timestamp)
	assert.True(t, diff < time.Minute && diff >= 0, msgAndArgs...)
}

// AssertNoError is a convenience wrapper for assert.NoError with better formatting
func AssertNoError(t *testing.T, err error, operation string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s failed: %v", operation, err)
	}
}

// AssertError checks that an error occurred and optionally matches a message
func AssertError(t *testing.T, err error, expectedMsg string, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Error(t, err, msgAndArgs...)
	if expectedMsg != "" {
		assert.Contains(t, err.Error(), expectedMsg, msgAndArgs...)
	}
}
```

**Step 4: Verify compilation**

Run: `cd backend && go build ./...`
Expected: Successful compilation

**Step 5: Commit**

```bash
git add backend/test/testutil/
git commit -m "feat: add test utilities for database, factories, and assertions"
```

---

### Task 6: Generate Mocks with Mockery

**Files:**
- Create: `backend/test/mocks/` (generated)

**Step 1: Generate mocks**

Run:
```bash
cd backend
mockery
```

Expected: Mocks generated in `backend/test/mocks/`

**Step 2: Verify mocks were created**

Run: `ls backend/test/mocks/`
Expected: Mock files for all interfaces (UserRepository, ConversationRepository, etc.)

**Step 3: Verify mock compilation**

Run: `cd backend && go build ./test/mocks/...`
Expected: Successful compilation

**Step 4: Commit**

```bash
git add backend/test/mocks/
git commit -m "feat: generate mocks for all repository and service interfaces"
```

---

### Task 7: Migrate Repository Tests

**Files:**
- Create: `backend/test/integration/repository_test.go`
- Delete: `backend/internal/repository/*_test.go`

**Step 1: Create consolidated repository test file**

Create `backend/test/integration/repository_test.go` by copying and consolidating all tests from:
- `backend/internal/repository/user_repository_test.go`
- `backend/internal/repository/conversation_repository_test.go`
- `backend/internal/repository/group_repository_test.go`
- `backend/internal/repository/message_repository_test.go`
- `backend/internal/repository/refresh_token_repository_test.go`
- `backend/internal/repository/integration_test.go`

Structure:
```go
package integration

import (
	"database/sql"
	"testing"

	"github.com/yourusername/virallens/backend/internal/repository"
	"github.com/yourusername/virallens/backend/test/testutil"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	// Setup runs before all tests
	// Note: Individual tests will call SetupTestDB
	os.Exit(m.Run())
}

// User Repository Tests
func TestUserRepository_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewUserRepository(db)
	user := testutil.CreateTestUser()

	err := repo.Create(context.Background(), user)
	testutil.AssertNoError(t, err, "Create user")
	testutil.AssertValidUUID(t, user.ID)
}

// ... (copy all other tests from individual files)
```

**Step 2: Update imports in migrated tests**

Update all imports to use:
- `github.com/yourusername/virallens/backend/test/testutil`
- `github.com/yourusername/virallens/backend/internal/repository`

**Step 3: Run migrated tests**

Run: `cd backend && go test ./test/integration/... -v`
Expected: All tests pass (same 48 + 5 = 53 repository tests)

**Step 4: Delete old test files**

Run:
```bash
rm backend/internal/repository/*_test.go
```

**Step 5: Commit**

```bash
git add backend/test/integration/ backend/internal/repository/
git commit -m "refactor: migrate repository tests to test/integration/"
```

---

### Task 8: Migrate Service Tests

**Files:**
- Create: `backend/test/unit/service_test.go`
- Delete: `backend/internal/service/*_test.go`

**Step 1: Create consolidated service test file**

Create `backend/test/unit/service_test.go` by copying and consolidating all tests from:
- `backend/internal/service/auth_service_test.go`
- `backend/internal/service/conversation_service_test.go`
- `backend/internal/service/group_service_test.go`
- `backend/internal/service/message_service_test.go`

Structure:
```go
package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourusername/virallens/backend/internal/service"
	"github.com/yourusername/virallens/backend/test/mocks"
	"github.com/yourusername/virallens/backend/test/testutil"
)

// Auth Service Tests
func TestAuthService_Register(t *testing.T) {
	userRepo := mocks.NewUserRepository(t)
	tokenRepo := mocks.NewRefreshTokenRepository(t)
	jwtService := mocks.NewJWTService(t)

	authService := service.NewAuthService(userRepo, tokenRepo, jwtService)

	testUser := testutil.CreateTestUser()
	
	userRepo.On("FindByEmail", mock.Anything, testUser.Email).Return(nil, domain.ErrUserNotFound)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
	jwtService.On("GenerateAccessToken", mock.Anything).Return("access-token", nil)
	jwtService.On("GenerateRefreshToken", mock.Anything).Return("refresh-token", nil)
	tokenRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

	result, err := authService.Register(context.Background(), testUser.Username, testUser.Email, "password123")
	
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "access-token", result.AccessToken)
}

// ... (copy all other tests from individual files)
```

**Step 2: Update mock usage**

Replace all manual mocks with generated mocks from `backend/test/mocks/`

**Step 3: Run migrated tests**

Run: `cd backend && go test ./test/unit/... -v`
Expected: All tests pass (same 46 service tests)

**Step 4: Delete old test files**

Run:
```bash
rm backend/internal/service/*_test.go
```

**Step 5: Commit**

```bash
git add backend/test/unit/ backend/internal/service/
git commit -m "refactor: migrate service tests to test/unit/ with generated mocks"
```

---

## Phase 3: Update Main Application with Wire

### Task 9: Rewrite main.go to Use Wire

**Files:**
- Modify: `backend/cmd/server/main.go`

**Step 1: Backup current main.go**

Run: `cp backend/cmd/server/main.go backend/cmd/server/main.go.backup`

**Step 2: Rewrite main.go**

Replace `backend/cmd/server/main.go` with:
```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/yourusername/virallens/backend/internal/api"
	"github.com/yourusername/virallens/backend/internal/config"
	"github.com/yourusername/virallens/backend/internal/wire"
)

func main() {
	// Load environment variables
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

	// Start WebSocket hub
	go app.WebSocketHub.Run()

	// Setup Echo server
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())

	// Routes
	api.SetupRoutes(e, app.AuthController, app.ConversationController, app.GroupController, app.JWTMiddleware, app.WebSocketHandler)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)

	// Graceful shutdown
	go func() {
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
```

**Step 3: Update api/routes.go**

Modify `backend/internal/api/routes.go` to accept Wire-injected controllers:
```go
package api

import (
	"github.com/labstack/echo/v4"
	"github.com/yourusername/virallens/backend/internal/middleware"
	"github.com/yourusername/virallens/backend/internal/websocket"
)

// SetupRoutes configures all API routes
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
	api.POST("/auth/register", authController.Register)
	api.POST("/auth/login", authController.Login)
	api.POST("/auth/refresh", authController.RefreshToken)

	// Protected routes
	protected := api.Group("")
	protected.Use(jwtMiddleware.Authenticate)

	// User routes
	protected.GET("/users/search", conversationController.SearchUsers)
	protected.GET("/users/me", authController.GetCurrentUser)

	// Conversation routes
	protected.GET("/conversations", conversationController.GetConversations)
	protected.GET("/conversations/:id", conversationController.GetConversation)
	protected.POST("/conversations", conversationController.CreateConversation)
	protected.GET("/conversations/:id/messages", conversationController.GetMessages)

	// Group routes
	protected.POST("/groups", groupController.CreateGroup)
	protected.GET("/groups/:id", groupController.GetGroup)
	protected.POST("/groups/:id/members", groupController.AddMember)
	protected.DELETE("/groups/:id/members/:userId", groupController.RemoveMember)
	protected.GET("/groups/:id/messages", groupController.GetMessages)

	// WebSocket
	protected.GET("/ws", wsHandler.HandleWebSocket)
}
```

**Step 4: Run the application**

Run: `cd backend && go run cmd/server/main.go`
Expected: Server starts successfully on configured port

**Step 5: Test basic endpoint**

Run (in another terminal): `curl http://localhost:8080/api/auth/login`
Expected: Response (even if error, proves server is running)

**Step 6: Stop server and commit**

```bash
git add backend/cmd/server/main.go backend/internal/api/routes.go
git commit -m "refactor: rewrite main.go to use Wire dependency injection"
```

---

## Phase 4: Monorepo Structure Setup

### Task 10: Create Root Package.json for npm Workspaces

**Files:**
- Create: `package.json`
- Modify: `frontend/package.json`

**Step 1: Create root package.json**

Create `package.json`:
```json
{
  "name": "virallens-monorepo",
  "version": "1.0.0",
  "private": true,
  "description": "ViralLens - Real-time chat application monorepo",
  "workspaces": [
    "frontend"
  ],
  "scripts": {
    "install:all": "npm install && cd backend && go mod download",
    "dev": "concurrently \"npm run dev:backend\" \"npm run dev:frontend\"",
    "dev:backend": "cd backend && go run cmd/server/main.go",
    "dev:frontend": "cd frontend && npm run dev",
    "build": "npm run build:backend && npm run build:frontend",
    "build:backend": "cd backend && go build -o bin/server cmd/server/main.go",
    "build:frontend": "cd frontend && npm run build",
    "test": "npm run test:backend && npm run test:frontend",
    "test:backend": "cd backend && go test ./...",
    "test:frontend": "cd frontend && npm run test",
    "clean": "npm run clean:backend && npm run clean:frontend && rm -rf node_modules",
    "clean:backend": "cd backend && rm -rf bin/",
    "clean:frontend": "cd frontend && rm -rf dist/ node_modules/"
  },
  "devDependencies": {
    "concurrently": "^8.2.2"
  },
  "engines": {
    "node": ">=18.0.0",
    "npm": ">=9.0.0"
  }
}
```

**Step 2: Install concurrently**

Run: `npm install`
Expected: concurrently installed

**Step 3: Test dev command**

Run: `npm run dev:backend` (in background)
Expected: Backend starts

Stop backend (Ctrl+C)

**Step 4: Commit**

```bash
git add package.json package-lock.json
git commit -m "feat: add root package.json with npm workspaces for monorepo"
```

---

### Task 11: Create Root Makefile

**Files:**
- Create: `Makefile`
- Create: `backend/Makefile`

**Step 1: Create backend Makefile**

Create `backend/Makefile`:
```makefile
.PHONY: help build run test clean wire mocks migrate

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the server binary
	go build -o bin/server cmd/server/main.go

run: ## Run the server
	go run cmd/server/main.go

test: ## Run all tests
	go test ./... -v -race -coverprofile=coverage.out

test-integration: ## Run integration tests only
	go test ./test/integration/... -v

test-unit: ## Run unit tests only
	go test ./test/unit/... -v

coverage: test ## Generate test coverage report
	go tool cover -html=coverage.out -o coverage.html

wire: ## Generate Wire dependency injection code
	cd internal/wire && wire

mocks: ## Generate mocks using Mockery
	mockery

clean: ## Clean build artifacts
	rm -rf bin/ coverage.out coverage.html

deps: ## Download dependencies
	go mod download
	go mod tidy

lint: ## Run linter
	golangci-lint run

fmt: ## Format code
	go fmt ./...

migrate-up: ## Run database migrations up
	migrate -path migrations -database "${DATABASE_URL}" up

migrate-down: ## Run database migrations down
	migrate -path migrations -database "${DATABASE_URL}" down

migrate-create: ## Create a new migration (use name=<migration_name>)
	migrate create -ext sql -dir migrations -seq $(name)

.DEFAULT_GOAL := help
```

**Step 2: Create root Makefile**

Create `Makefile`:
```makefile
.PHONY: help install dev prod test clean docker-up docker-down

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install: ## Install all dependencies
	@echo "Installing dependencies..."
	npm install
	cd backend && go mod download

dev: ## Start development servers (backend + frontend)
	npm run dev

dev-backend: ## Start backend only
	cd backend && make run

dev-frontend: ## Start frontend only
	cd frontend && npm run dev

build: ## Build both backend and frontend
	@echo "Building backend..."
	cd backend && make build
	@echo "Building frontend..."
	cd frontend && npm run build

test: ## Run all tests
	@echo "Running backend tests..."
	cd backend && make test
	@echo "Running frontend tests..."
	cd frontend && npm run test

test-backend: ## Run backend tests only
	cd backend && make test

test-frontend: ## Run frontend tests only
	cd frontend && npm run test

clean: ## Clean all build artifacts
	@echo "Cleaning backend..."
	cd backend && make clean
	@echo "Cleaning frontend..."
	cd frontend && rm -rf dist/ node_modules/
	rm -rf node_modules/

docker-up: ## Start Docker containers
	docker-compose up -d

docker-down: ## Stop Docker containers
	docker-compose down

docker-logs: ## Show Docker logs
	docker-compose logs -f

wire: ## Generate Wire dependency injection code
	cd backend && make wire

mocks: ## Generate mocks
	cd backend && make mocks

.DEFAULT_GOAL := help
```

**Step 3: Test Makefiles**

Run: `make help`
Expected: Shows available commands

Run: `cd backend && make help`
Expected: Shows backend commands

**Step 4: Commit**

```bash
git add Makefile backend/Makefile
git commit -m "feat: add Makefiles for monorepo and backend task automation"
```

---

### Task 12: Update .env.example

**Files:**
- Modify: `backend/.env.example`

**Step 1: Update .env.example with all new config variables**

Replace `backend/.env.example` with:
```env
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
SERVER_READ_TIMEOUT=15s
SERVER_WRITE_TIMEOUT=15s
SERVER_SHUTDOWN_TIMEOUT=30s

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password_here
DB_NAME=virallens
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# Test Database (for running tests)
TEST_DATABASE_URL=postgres://postgres:your_password_here@localhost:5432/virallens_test?sslmode=disable

# JWT Configuration
JWT_ACCESS_SECRET=your_super_secret_access_key_change_this_in_production
JWT_REFRESH_SECRET=your_super_secret_refresh_key_change_this_in_production
JWT_ACCESS_EXPIRATION=15m
JWT_REFRESH_EXPIRATION=168h

# Application Configuration
APP_ENV=development
LOG_LEVEL=info

# Migration Configuration
DATABASE_URL=postgres://postgres:your_password_here@localhost:5432/virallens?sslmode=disable
```

**Step 2: Commit**

```bash
git add backend/.env.example
git commit -m "docs: update .env.example with all configuration variables"
```

---

### Task 13: Enhance docker-compose.yml

**Files:**
- Modify: `docker-compose.yml`

**Step 1: Check if docker-compose.yml exists**

Run: `ls docker-compose.yml`
Expected: File exists OR file not found

**Step 2: Create or update docker-compose.yml**

Create/replace `docker-compose.yml`:
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: virallens-db
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: virallens
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backend/migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  postgres-test:
    image: postgres:15-alpine
    container_name: virallens-test-db
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: virallens_test
    ports:
      - "5433:5432"
    tmpfs:
      - /var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: virallens-backend
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      SERVER_PORT: 8080
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: postgres
      DB_NAME: virallens
      JWT_ACCESS_SECRET: dev_access_secret_change_in_production
      JWT_REFRESH_SECRET: dev_refresh_secret_change_in_production
    ports:
      - "8080:8080"
    volumes:
      - ./backend:/app
    command: go run cmd/server/main.go

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    container_name: virallens-frontend
    depends_on:
      - backend
    ports:
      - "5173:5173"
    volumes:
      - ./frontend:/app
      - /app/node_modules
    environment:
      VITE_API_URL: http://localhost:8080
      VITE_WS_URL: ws://localhost:8080
    command: npm run dev

volumes:
  postgres_data:
```

**Step 3: Test docker-compose**

Run: `docker-compose config`
Expected: Valid YAML, no errors

**Step 4: Commit**

```bash
git add docker-compose.yml
git commit -m "feat: add comprehensive docker-compose.yml for development environment"
```

---

### Task 14: Create Backend Dockerfile

**Files:**
- Create: `backend/Dockerfile`
- Create: `backend/.dockerignore`

**Step 1: Create Dockerfile**

Create `backend/Dockerfile`:
```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server cmd/server/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./server"]
```

**Step 2: Create .dockerignore**

Create `backend/.dockerignore`:
```
# Binaries
bin/
*.exe
*.dll
*.so
*.dylib

# Test files
*.test
*.out
coverage.out
coverage.html

# Development files
.env
.env.local
.DS_Store
*.swp
*.swo
*~

# IDE
.vscode/
.idea/

# Git
.git/
.gitignore

# Documentation
docs/
README.md

# Test database
*.db
```

**Step 3: Test Dockerfile build**

Run: `cd backend && docker build -t virallens-backend:test .`
Expected: Build succeeds

**Step 4: Commit**

```bash
git add backend/Dockerfile backend/.dockerignore
git commit -m "feat: add Dockerfile and .dockerignore for backend"
```

---

## Phase 5: Documentation and Final Touches

### Task 15: Update Backend README

**Files:**
- Create: `backend/README.md`

**Step 1: Create comprehensive backend README**

Create `backend/README.md`:
```markdown
# ViralLens Backend

Real-time chat application backend built with Go, Echo, PostgreSQL, and WebSockets.

## Architecture

This backend follows Clean Architecture principles with dependency injection via Google Wire:

- **Domain Layer**: Business entities and interfaces (`internal/domain/`)
- **Repository Layer**: Data persistence implementations (`internal/repository/`)
- **Service Layer**: Business logic (`internal/service/`)
- **API Layer**: HTTP controllers (`internal/api/`)
- **WebSocket Layer**: Real-time messaging (`internal/websocket/`)

### Dependency Injection

We use [Google Wire](https://github.com/google/wire/wire) for compile-time dependency injection. All dependencies are wired together in `internal/wire/`.

## Tech Stack

- **Language**: Go 1.22+
- **Web Framework**: Echo v4
- **Database**: PostgreSQL 15+
- **WebSocket**: gorilla/websocket
- **Authentication**: JWT
- **DI**: Google Wire
- **Testing**: testify, Mockery
- **Migrations**: golang-migrate

## Getting Started

### Prerequisites

- Go 1.22 or higher
- PostgreSQL 15+
- Make
- Wire CLI: `go install github.com/google/wire/cmd/wire@latest`
- Mockery: `go install github.com/vektra/mockery/v2@latest`

### Installation

1. **Install dependencies**:
   ```bash
   make deps
   ```

2. **Setup environment**:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Setup database**:
   ```bash
   # Create database
   createdb virallens
   
   # Run migrations
   make migrate-up
   ```

4. **Generate Wire code** (if not already generated):
   ```bash
   make wire
   ```

5. **Run the server**:
   ```bash
   make run
   ```

The server will start on `http://localhost:8080`

## Development

### Project Structure

```
backend/
├── cmd/server/           # Application entry point
├── internal/
│   ├── api/             # HTTP controllers
│   ├── config/          # Configuration management
│   ├── domain/          # Business entities & interfaces
│   ├── middleware/      # HTTP middleware
│   ├── repository/      # Data persistence layer
│   ├── service/         # Business logic layer
│   ├── websocket/       # WebSocket handlers
│   └── wire/            # Dependency injection setup
├── test/
│   ├── integration/     # Integration tests
│   ├── unit/           # Unit tests
│   ├── mocks/          # Generated mocks
│   └── testutil/       # Test utilities
├── migrations/          # Database migrations
└── Makefile            # Build automation
```

### Available Commands

```bash
make help              # Show all available commands
make build             # Build the server binary
make run               # Run the server
make test              # Run all tests
make test-integration  # Run integration tests only
make test-unit         # Run unit tests only
make coverage          # Generate coverage report
make wire              # Generate Wire DI code
make mocks             # Generate mocks
make clean             # Clean build artifacts
make migrate-up        # Run migrations up
make migrate-down      # Run migrations down
```

### Running Tests

```bash
# All tests
make test

# Integration tests (require test database)
make test-integration

# Unit tests (use mocks, no database required)
make test-unit

# With coverage
make coverage
```

### Database Migrations

```bash
# Run migrations
make migrate-up

# Rollback migrations
make migrate-down

# Create new migration
make migrate-create name=add_users_table
```

### Generating Code

```bash
# Regenerate Wire dependency injection
make wire

# Regenerate mocks
make mocks
```

## Configuration

Configuration is loaded from environment variables. See `.env.example` for all available options.

Key configuration areas:
- **Server**: Port, host, timeouts
- **Database**: Connection settings, pool configuration
- **JWT**: Secrets and expiration times
- **App**: Environment, log level

## Testing

Tests are organized into:

- **Unit Tests** (`test/unit/`): Test services with mocked dependencies
- **Integration Tests** (`test/integration/`): Test repositories with real database
- **Test Utilities** (`test/testutil/`): Shared test helpers and factories

### Test Database Setup

```bash
# Create test database
createdb virallens_test

# Set TEST_DATABASE_URL in .env
TEST_DATABASE_URL=postgres://postgres:password@localhost:5432/virallens_test?sslmode=disable
```

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login user
- `POST /api/auth/refresh` - Refresh access token

### Users
- `GET /api/users/me` - Get current user
- `GET /api/users/search` - Search users

### Conversations
- `GET /api/conversations` - List conversations
- `GET /api/conversations/:id` - Get conversation details
- `POST /api/conversations` - Create conversation
- `GET /api/conversations/:id/messages` - Get messages

### Groups
- `POST /api/groups` - Create group
- `GET /api/groups/:id` - Get group details
- `POST /api/groups/:id/members` - Add member
- `DELETE /api/groups/:id/members/:userId` - Remove member
- `GET /api/groups/:id/messages` - Get group messages

### WebSocket
- `GET /api/ws` - WebSocket connection (authenticated)

## Deployment

### Using Docker

```bash
# Build image
docker build -t virallens-backend .

# Run container
docker run -p 8080:8080 --env-file .env virallens-backend
```

### Using Docker Compose (from root)

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f backend
```

## Contributing

1. Create a feature branch
2. Write tests first (TDD)
3. Implement feature
4. Ensure all tests pass: `make test`
5. Regenerate mocks if interfaces changed: `make mocks`
6. Regenerate Wire if providers changed: `make wire`
7. Submit pull request

## License

MIT
```

**Step 2: Commit**

```bash
git add backend/README.md
git commit -m "docs: add comprehensive backend README"
```

---

### Task 16: Update Root README

**Files:**
- Modify: `README.md` (or create if doesn't exist)

**Step 1: Create/update root README**

Create `README.md`:
```markdown
# ViralLens

A modern, real-time chat application built with Go and React.

## Features

- 💬 Real-time messaging via WebSocket
- 👥 One-on-one conversations
- 🎭 Group chats
- 🔐 JWT authentication
- 📱 Responsive UI with HeroUI components
- 🎨 Clean Architecture backend
- ⚡ Fast and scalable

## Tech Stack

### Backend
- **Language**: Go 1.22+
- **Framework**: Echo v4
- **Database**: PostgreSQL 15+
- **WebSocket**: gorilla/websocket
- **DI**: Google Wire
- **Testing**: testify + Mockery

### Frontend
- **Framework**: React 19
- **Language**: TypeScript
- **State**: Zustand
- **UI**: HeroUI (NextUI)
- **Build**: Vite

## Quick Start

### Prerequisites

- Go 1.22+
- Node.js 18+
- PostgreSQL 15+
- Make

### Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/yourusername/virallens.git
   cd virallens
   ```

2. **Install all dependencies**:
   ```bash
   make install
   ```

3. **Setup environment**:
   ```bash
   cp backend/.env.example backend/.env
   # Edit backend/.env with your configuration
   ```

4. **Start database** (using Docker):
   ```bash
   make docker-up
   ```

5. **Run migrations**:
   ```bash
   cd backend && make migrate-up
   ```

6. **Start development servers**:
   ```bash
   make dev
   ```

The application will be available at:
- Frontend: http://localhost:5173
- Backend: http://localhost:8080

## Development

### Project Structure

```
virallens/
├── backend/              # Go backend
│   ├── cmd/             # Application entry points
│   ├── internal/        # Internal packages
│   ├── test/            # Tests and test utilities
│   └── migrations/      # Database migrations
├── frontend/            # React frontend
│   ├── src/            # Source code
│   │   ├── components/ # React components
│   │   ├── pages/      # Page components
│   │   ├── stores/     # Zustand stores
│   │   └── services/   # API & WebSocket services
│   └── public/         # Static assets
├── docs/                # Documentation
└── scripts/             # Utility scripts
```

### Available Commands

#### Root Commands
```bash
make help          # Show available commands
make install       # Install all dependencies
make dev           # Start backend + frontend
make dev-backend   # Start backend only
make dev-frontend  # Start frontend only
make build         # Build both projects
make test          # Run all tests
make clean         # Clean build artifacts
make docker-up     # Start Docker services
make docker-down   # Stop Docker services
```

#### Backend Commands
```bash
cd backend
make help              # Show backend commands
make build             # Build server binary
make run               # Run server
make test              # Run tests
make wire              # Generate Wire DI code
make mocks             # Generate mocks
make migrate-up        # Run migrations
```

#### Frontend Commands
```bash
cd frontend
npm run dev            # Start dev server
npm run build          # Build for production
npm run preview        # Preview production build
npm run test           # Run tests
npm run lint           # Lint code
```

### Running Tests

```bash
# All tests (backend + frontend)
make test

# Backend only
make test-backend

# Frontend only
make test-frontend
```

### Database Migrations

```bash
# Run migrations
cd backend && make migrate-up

# Rollback migrations
cd backend && make migrate-down

# Create new migration
cd backend && make migrate-create name=your_migration_name
```

## Architecture

### Backend (Clean Architecture)

The backend follows Clean Architecture with clear separation of concerns:

```
Domain Layer (entities, interfaces)
    ↓
Repository Layer (data persistence)
    ↓
Service Layer (business logic)
    ↓
API Layer (HTTP controllers)
```

Dependencies are managed with **Google Wire** for compile-time dependency injection.

### Frontend (Component-Based)

The frontend uses React with functional components and hooks:

```
Pages → Components → Services (API/WebSocket) → Stores (Zustand)
```

## Deployment

### Docker Compose (Development)

```bash
make docker-up
```

### Production

See individual README files:
- Backend: [backend/README.md](backend/README.md)
- Frontend: [frontend/README.md](frontend/README.md)

## API Documentation

### Authentication
- `POST /api/auth/register` - Register
- `POST /api/auth/login` - Login
- `POST /api/auth/refresh` - Refresh token

### Conversations
- `GET /api/conversations` - List conversations
- `POST /api/conversations` - Create conversation
- `GET /api/conversations/:id/messages` - Get messages

### Groups
- `POST /api/groups` - Create group
- `POST /api/groups/:id/members` - Add member
- `GET /api/groups/:id/messages` - Get messages

### WebSocket
- `GET /api/ws` - WebSocket connection

## Contributing

1. Fork the repository
2. Create your feature branch
3. Write tests
4. Implement your feature
5. Ensure tests pass
6. Submit a pull request

## License

MIT

## Support

For issues and questions, please open an issue on GitHub.
```

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add comprehensive root README for monorepo"
```

---

### Task 17: Run Full Test Suite

**Files:**
- N/A (verification step)

**Step 1: Run all backend tests**

Run: `cd backend && go test ./... -v`
Expected: All tests pass (99+ tests)

**Step 2: Run integration tests specifically**

Run: `cd backend && go test ./test/integration/... -v`
Expected: All integration tests pass

**Step 3: Run unit tests specifically**

Run: `cd backend && go test ./test/unit/... -v`
Expected: All unit tests pass

**Step 4: Check test coverage**

Run: `cd backend && make coverage`
Expected: Coverage report generated, opens in browser

**Step 5: Document results**

Note: No commit needed, this is verification only

---

### Task 18: Verify Application Startup

**Files:**
- N/A (verification step)

**Step 1: Start backend**

Run: `cd backend && make run`
Expected: Server starts on port 8080, no errors

**Step 2: Test health check**

Run (in another terminal): `curl http://localhost:8080/api/auth/login -X POST -H "Content-Type: application/json" -d '{"email":"test@example.com","password":"test"}'`
Expected: JSON response (even if error, proves server is responding)

**Step 3: Stop backend**

Press Ctrl+C

**Step 4: Test monorepo dev command**

Run: `make dev`
Expected: Both backend and frontend start, no errors

**Step 5: Verify frontend loads**

Open browser: http://localhost:5173
Expected: Frontend loads successfully

**Step 6: Stop all services**

Press Ctrl+C

---

### Task 19: Final Verification Checklist

**Files:**
- N/A (verification step)

**Step 1: Verify Wire generation works**

Run:
```bash
cd backend/internal/wire
wire
echo $?
```
Expected: Exit code 0, wire_gen.go updated or unchanged

**Step 2: Verify mock generation works**

Run:
```bash
cd backend
mockery
```
Expected: Mocks generated or already up to date

**Step 3: Verify all Makefile commands**

Run:
```bash
make help
cd backend && make help
```
Expected: Help output shows all commands

**Step 4: Verify Docker Compose**

Run:
```bash
docker-compose config
```
Expected: Valid YAML configuration

**Step 5: Verify npm scripts**

Run:
```bash
npm run
```
Expected: Shows all available scripts

**Step 6: Create verification checklist document**

Create `docs/VERIFICATION.md`:
```markdown
# Implementation Verification Checklist

## Phase 1: Wire DI ✅
- [x] Wire and Mockery dependencies installed
- [x] Config package created with validation
- [x] Repository structs made public with Impl suffix
- [x] Wire configuration created
- [x] wire_gen.go generated successfully

## Phase 2: Test Infrastructure ✅
- [x] Test utilities created (db, factory, assert)
- [x] Mocks generated with Mockery
- [x] Repository tests migrated to test/integration/
- [x] Service tests migrated to test/unit/

## Phase 3: Main Application ✅
- [x] main.go rewritten to use Wire
- [x] Routes updated to accept Wire-injected controllers
- [x] Application starts successfully
- [x] All endpoints responding

## Phase 4: Monorepo Structure ✅
- [x] Root package.json with npm workspaces
- [x] Root Makefile created
- [x] Backend Makefile created
- [x] .env.example updated
- [x] docker-compose.yml enhanced
- [x] Backend Dockerfile created

## Phase 5: Documentation ✅
- [x] Backend README created
- [x] Root README updated
- [x] All tests passing
- [x] Application verified working

## Test Results
- Total tests: 99+
- Integration tests: 53
- Unit tests: 46
- Coverage: [Check coverage.html]

## Commands Verified
- `make dev` - ✅ Starts both servers
- `make test` - ✅ All tests pass
- `make build` - ✅ Builds both projects
- `make wire` - ✅ Generates Wire code
- `make mocks` - ✅ Generates mocks
- `docker-compose up` - ✅ Starts services

## Breaking Changes Applied
- ✅ Repository structs now public (userRepository → UserRepositoryImpl)
- ✅ Test files moved from internal/*/test.go to test/
- ✅ Config loading changed from direct env vars to Config struct
- ✅ main.go completely rewritten
- ✅ Import paths updated throughout

## Known Issues
None - all functionality preserved, tests passing

## Next Steps
1. Deploy to staging
2. Run smoke tests
3. Monitor performance
4. Update documentation as needed
```

**Step 7: Commit verification document**

```bash
git add docs/VERIFICATION.md
git commit -m "docs: add implementation verification checklist"
```

---

## Summary

This plan implements:

1. **Wire Dependency Injection**: Clean compile-time DI with generated code
2. **Improved Test Structure**: Organized tests in `test/` with utilities and mocks
3. **Enhanced Configuration**: Type-safe config with validation
4. **Monorepo Setup**: npm workspaces + Makefiles for easy development
5. **Better Documentation**: Comprehensive READMEs and guides

**Total Tasks**: 19
**Estimated Time**: 8-12 hours
**Breaking Changes**: Yes, but all functionality preserved
**Test Coverage**: Maintained (99+ tests all passing)

## Post-Implementation

After completing all tasks:

1. Run full test suite one more time
2. Test all Makefile commands
3. Verify Docker Compose works
4. Test monorepo dev workflow
5. Review generated Wire code
6. Check all documentation
7. Create a new feature branch for future work

## Notes

- Wire code must be regenerated whenever providers change
- Mocks must be regenerated when interfaces change
- Always run tests after Wire/mock regeneration
- Docker Compose requires Docker installed and running
- Test database must be created before running integration tests
