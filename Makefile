.PHONY: help install dev dev-backend dev-frontend build test test-backend test-frontend clean docker-up docker-down docker-logs wire mocks

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
