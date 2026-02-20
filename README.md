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
