# ViralLens

A modern, real-time chat application built with Go and React.

## Features

- 💬 **Real-time messaging**: Seamless instant messaging via WebSocket.
- 👥 **One-on-one & Group chats**: Private conversations and collaborative group spaces.
- 🔐 **JWT authentication**: Secure access with access and refresh tokens.
- 🛑 **Rate Limiting**: Built-in protection against message spam.
- 📱 **Responsive UI**: Sleek, mobile-friendly design using HeroUI (NextUI) components.
- ⚡ **Go-powered Backend**: High-performance API built with Gin.
- 🚀 **CI/CD**: Automated building and testing with GitHub Actions.

## Tech Stack

### Backend

- **Language**: Go 1.25+
- **Framework**: [Gin](https://gin-gonic.com/)
- **Database**: [PostgreSQL](https://www.postgresql.org/) (via [GORM](https://gorm.io/))
- **WebSocket**: [gorilla/websocket](https://github.com/gorilla/websocket)
- **DI**: [Google Wire](https://github.com/google/wire)
- **Configuration**: [Viper](https://github.com/spf13/viper)

### Frontend

- **Framework**: [React 19](https://react.dev/)
- **Build Tool**: [Vite](https://vitejs.dev/)
- **Language**: TypeScript
- **State Management**: [Zustand](https://github.com/pmndrs/zustand)
- **UI Library**: [HeroUI](https://heroui.com/) (formerly NextUI)
- **Styling**: Tailwind CSS

---

## Quick Start

### Prerequisites

- [Go](https://go.dev/doc/install) (1.25+)
- [Node.js](https://nodejs.org/) (20+)
- [Docker](https://www.docker.com/) (for PostgreSQL)
- [Make](https://www.gnu.org/software/make/)

### Installation & Setup

1. **Clone & Install Dependencies**

   ```bash
   git clone https://github.com/yourusername/virallens.git
   cd virallens
   make install
   ```

2. **Environment Configuration**

   ```bash
   cp backend/.env.example backend/.env
   # Update backend/.env with your settings (DB_URL, JWT_SECRET, etc.)
   ```

3. **Infrastructure & Database**

   ```bash
   docker-compose up -d  # Start PostgreSQL
   cd backend && make migrate-up
   ```

4. **Run Development Mode**
   ```bash
   make dev  # Starts both Backend and Frontend
   ```

- **Frontend**: [http://localhost:5173](http://localhost:5173)
- **Backend API**: [http://localhost:8080](http://localhost:8080)

---

## Architecture Overview

### Backend (Layered Architecture)

The backend is structured into modular components using a layered approach to ensure separation of concerns:

- **Routes**: Defines the API endpoints and connects them to controllers and middlewares.
- **Controllers**: Handle HTTP requests, validate input, and call service methods.
- **Services**: Contain business logic and orchestrate domain operations.
- **Repositories**: Abstract database access using GORM.
- **Middlewares**: Handle cross-cutting concerns like Auth (JWT) and Rate Limiting.
- **WebSocket**: A dedicated handler for managing real-time connections and message broadcasting.

### Frontend (Modern React Hooks & State)

- **Components**: Atomic and reusable UI elements.
- **Hooks**: Custom logic for API calls and WebSocket synchronization.
- **Zustand Stores**: Centralized state for user sessions, chat history, and UI state.
- **API/WS Layer**: Service-based communication with the backend.

---

## Key Trade-offs & Design Decisions

- **In-Memory Rate Limiting**: We implemented a per-user in-memory rate limiter for messages.
  - _Trade-off_: While fast and simple to implement, it doesn't persist across restarts or scale horizontally in a multi-instance environment (for which Redis would be preferred).
- **Monolithic Repository**: Both frontend and backend reside in a single repository.
  - _Trade-off_: Simplifies development synchronization and CI/CD but may become harder to manage as the team and codebase grow significantly.
- **GORM vs. Raw SQL**: Used GORM for rapid development and ease of migration management.
  - _Trade-off_: Abstracted away some performance optimizations possible with raw SQL for the sake of developer productivity.
- **JWT-based Auth**: Stateless authentication using JWTs.
  - _Trade-off_: Easier to scale than session-based auth, but requires careful token management (refresh tokens) and lacks immediate revocation without an blocklist.
- **Zustand over Redux**: Chosen for its minimal boilerplate and ease of use in smaller to medium applications.
  - _Trade-off_: Less prescriptive than Redux, which requires discipline to maintain a clean state structure.

---

## CI/CD Pipeline

We use **GitHub Actions** to maintain code quality:

- **Backend**: Automated `go build` and `go test` (with race detection) on every pull request.
- **Frontend**: Automated `npm run build` and `vitest` to ensure UI stability.

---

## License

MIT
