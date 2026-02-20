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
