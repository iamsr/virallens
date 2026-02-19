# ViralLens - Real-Time Chat Application

Production-ready real-time chat application with one-to-one and group messaging.

## Features

- 🔐 JWT authentication with refresh tokens
- 💬 One-to-one private messaging
- 👥 Group chat creation and management
- ⚡ Real-time WebSocket communication
- 📜 Message history with cursor-based pagination
- 🛡️ Rate limiting and spam protection

## Tech Stack

**Backend:** Go 1.21+, Echo, PostgreSQL, WebSocket, Wire DI
**Frontend:** React 18, TypeScript, Zustand, HeroUI
**Infrastructure:** Docker, Docker Compose, GitHub Actions

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.21+ (for local development)
- Node.js 18+ (for local development)

### Run with Docker

```bash
docker-compose up --build
```

- Frontend: http://localhost:5173
- Backend API: http://localhost:8080
- WebSocket: ws://localhost:8080/ws

### Local Development

**Backend:**
```bash
cd backend
cp .env.example .env
go run cmd/server/main.go
```

**Frontend:**
```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

## Testing

**Backend:**
```bash
cd backend
go test ./... -v                    # Unit tests
go test ./tests/integration/... -v  # Integration tests
```

**Frontend:**
```bash
cd frontend
npm test                    # Unit tests
npm run test:e2e           # E2E tests
```

## API Documentation

See [docs/API.md](./docs/API.md)

## Architecture

See [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)

## License

MIT
