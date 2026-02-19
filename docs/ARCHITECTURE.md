# Architecture Documentation

## Overview

ViralLens follows **Clean Architecture** principles with clear separation of concerns and dependency inversion. The backend is structured in layers, with each layer having specific responsibilities.

---

## Clean Architecture Layers

```
┌─────────────────────────────────────────┐
│         Presentation Layer              │
│    (Controllers, WebSocket Handlers)    │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│          Service Layer                  │
│      (Business Logic)                   │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│        Repository Layer                 │
│      (Data Access)                      │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│         Domain Layer                    │
│    (Entities & Interfaces)              │
└─────────────────────────────────────────┘
```

### 1. Domain Layer (`internal/domain/`)
**Responsibility:** Define core business entities and repository interfaces.

- Pure Go structs representing business entities
- Repository interfaces (dependency inversion)
- Domain-specific errors
- No external dependencies

**Example:**
```go
type User struct {
    ID       uuid.UUID
    Username string
    Email    string
}

type UserRepository interface {
    Create(user *User) error
    GetByID(id uuid.UUID) (*User, error)
}
```

---

### 2. Repository Layer (`internal/repository/`)
**Responsibility:** Implement data access logic.

- Implements domain repository interfaces
- Direct database interaction
- SQL queries and transactions
- Error handling and mapping to domain errors

**Key Features:**
- PostgreSQL with `database/sql`
- Transaction support for complex operations
- Efficient indexing usage
- Connection pooling

---

### 3. Service Layer (`internal/service/`)
**Responsibility:** Implement business logic.

- Orchestrates repository operations
- Enforces business rules
- Authorization checks
- Input validation
- Uses repository interfaces (testable via mocking)

**Example:**
```go
type AuthService struct {
    userRepo UserRepository
    tokenRepo RefreshTokenRepository
    jwtService JWTService
}

func (s *AuthService) Login(username, password string) (*User, string, error) {
    // Business logic here
}
```

---

### 4. Presentation Layer (`internal/controller/`, `internal/websocket/`)
**Responsibility:** Handle HTTP requests and WebSocket connections.

- Echo framework for REST API
- Request validation
- Response formatting
- HTTP status codes
- Calls service layer methods

---

## Dependency Flow

```
Controller → Service → Repository → Domain
     ↑          ↑          ↑
     └──────────┴──────────┴────── Interfaces defined in Domain
```

**Key Principle:** Dependencies point inward. Outer layers depend on inner layers, never the reverse.

---

## WebSocket Architecture

### Hub Pattern

The WebSocket implementation uses a **Hub pattern** for managing client connections:

```
┌─────────────────────────────────────────┐
│              WebSocket Hub              │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │   Connection Registry           │   │
│  │   map[userID]*Client            │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │   Broadcast Channels            │   │
│  │   - Register                    │   │
│  │   - Unregister                  │   │
│  │   - Broadcast                   │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
         │         │         │
    ┌────┴───┐ ┌──┴───┐ ┌──┴────┐
    │ Client │ │Client│ │Client │
    │   1    │ │  2   │ │  3    │
    └────────┘ └──────┘ └───────┘
```

### Client Management

Each WebSocket client has:
- **Read Pump:** Goroutine reading messages from client
- **Write Pump:** Goroutine writing messages to client
- **Message Buffer:** Channel for outgoing messages
- **Ping/Pong:** Heartbeat mechanism

### Message Flow

1. Client sends message via WebSocket
2. Read pump receives and validates
3. Message routed to service layer
4. Service processes and saves to database
5. Hub broadcasts to relevant participants
6. Write pumps send to connected clients

---

## Database Design

### Schema Overview

```
users ──────┐
            │
            ├─→ conversations ←─→ conversation_participants
            │
            ├─→ groups ←─→ group_members
            │
            ├─→ messages (conversation_id OR group_id)
            │
            └─→ refresh_tokens
```

### Key Design Decisions

1. **Unified Messages Table**
   - Single table for both conversation and group messages
   - CHECK constraint ensures message belongs to exactly one target
   - Simplifies message history queries

2. **Many-to-Many Relationships**
   - `conversation_participants`: Users ↔ Conversations
   - `group_members`: Users ↔ Groups
   - Allows efficient membership queries

3. **Cursor-Based Pagination**
   - Uses `created_at` timestamp as cursor
   - Indexed DESC for efficient "newest first" queries
   - Consistent pagination even with new messages

4. **Indexes for Performance**
   - Composite indexes on message queries
   - Foreign key indexes
   - Unique constraints on usernames/emails

---

## Authentication Flow

### JWT with Refresh Tokens

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ 1. Login (username, password)
       ▼
┌─────────────────────────────────┐
│   Auth Service                  │
│                                 │
│  - Verify credentials           │
│  - Generate access token (15m)  │
│  - Generate refresh token (7d)  │
│  - Store refresh token in DB    │
└──────┬──────────────────────────┘
       │ 2. Return both tokens
       ▼
┌─────────────┐
│   Client    │
│  - Store tokens in localStorage │
│  - Use access token for API     │
└──────┬──────────────────────────┘
       │ 3. Access token expires
       │ 4. Call /auth/refresh
       ▼
┌─────────────────────────────────┐
│   Auth Service                  │
│  - Validate refresh token       │
│  - Generate new access token    │
└──────┬──────────────────────────┘
       │ 5. Return new access token
       ▼
┌─────────────┐
│   Client    │
│  - Continue with new token      │
└─────────────┘
```

### Security Features

- Access tokens: Short-lived (15 minutes)
- Refresh tokens: Longer-lived (7 days), stored in database
- Refresh tokens can be revoked (logout)
- Passwords hashed with bcrypt
- Token validation on every protected request

---

## Rate Limiting Strategy

### Implementation

Uses `golang.org/x/time/rate` with token bucket algorithm.

**Per-User Rate Limiting:**
- 10 requests per minute per user
- Applied to message sending endpoints
- Returns 429 Too Many Requests when exceeded

**Middleware:**
```go
type RateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
}
```

### Why Per-User?

- Prevents spam from individual users
- Doesn't penalize legitimate traffic
- Simple to implement and reason about

---

## Dependency Injection

### Wire

Using Google Wire for compile-time dependency injection:

**Benefits:**
- Type-safe
- No reflection overhead
- Clear dependency graph
- Easy to test (mock injection)

**Provider Pattern:**
```go
// wire.go
func InitializeServer() (*echo.Echo, error) {
    wire.Build(
        config.NewDatabase,
        repository.NewUserRepository,
        service.NewAuthService,
        controller.NewAuthController,
        // ... more providers
    )
    return nil, nil
}
```

**Generated Code:**
```bash
cd backend
wire  # Generates wire_gen.go
```

---

## Error Handling

### Domain Errors

Defined in `internal/domain/errors.go`:
```go
var (
    ErrUserNotFound = errors.New("user not found")
    ErrUnauthorized = errors.New("unauthorized")
    // ...
)
```

### Error Flow

1. Repository returns domain error or wrapped error
2. Service checks error type, returns appropriate response
3. Controller maps to HTTP status code
4. Client receives standardized error response

### HTTP Status Mapping

- `ErrNotFound` → 404
- `ErrUnauthorized` → 401
- `ErrForbidden` → 403
- `ErrAlreadyExists` → 409
- `ErrValidation` → 400
- Other → 500

---

## Testing Strategy

### Unit Tests
- **Repository:** Test database operations with test DB
- **Service:** Test business logic with mocked repositories
- **Controller:** Test HTTP handling with mocked services

### Integration Tests
- Full API flow tests
- Real database (test instance)
- WebSocket connection tests
- Authorization tests

### E2E Tests
- Complete user journeys
- Frontend + Backend integration
- Multi-client scenarios
- Real-time message delivery

---

## Frontend Architecture

### State Management (Zustand)

```
┌─────────────────────────────────────┐
│         Zustand Stores              │
│                                     │
│  ┌──────────────┐  ┌─────────────┐ │
│  │  Auth Store  │  │ Chat Store  │ │
│  └──────────────┘  └─────────────┘ │
│                                     │
│  ┌──────────────┐  ┌─────────────┐ │
│  │   WS Store   │  │  UI Store   │ │
│  └──────────────┘  └─────────────┘ │
└─────────────────────────────────────┘
         ▲                    │
         │                    ▼
    ┌────────────────────────────┐
    │   React Components         │
    └────────────────────────────┘
```

### Service Layer

- **API Service:** Axios with interceptors for token management
- **WebSocket Service:** Connection management, auto-reconnect
- Type-safe with TypeScript

### Component Structure

```
pages/
  ├── Login.tsx
  ├── Register.tsx
  └── Chat.tsx
components/
  ├── auth/
  ├── chat/
  └── common/
```

---

## Production Considerations

### Scalability

**Current Architecture Supports:**
- Horizontal scaling of API servers (stateless)
- Database connection pooling
- WebSocket connections per server

**For Large Scale:**
- Redis for WebSocket message distribution
- Database read replicas
- Message queue for async operations
- CDN for frontend assets

### Monitoring

**Recommended:**
- Application logs (structured logging)
- Database query performance monitoring
- WebSocket connection metrics
- Error tracking (Sentry)
- Health check endpoints

### Security

- Environment variables for secrets
- HTTPS in production
- CORS configuration
- SQL injection prevention (parameterized queries)
- XSS prevention (React auto-escaping)
- Rate limiting
- Token expiration and refresh

---

## Technology Choices

### Why Go?
- Excellent concurrency (goroutines)
- Strong typing and compile-time checks
- Great performance for WebSocket handling
- Simple deployment (single binary)

### Why PostgreSQL?
- ACID compliance
- Rich indexing options
- UUID support
- JSON support (future extensibility)
- Battle-tested reliability

### Why Echo?
- Lightweight and fast
- Good middleware support
- Simple routing
- WebSocket support

### Why Zustand?
- Simple API
- No boilerplate
- React hooks integration
- Good TypeScript support
- Lightweight (small bundle)

### Why HeroUI?
- Modern component library
- Good TypeScript support
- Tailwind integration
- Accessible components
