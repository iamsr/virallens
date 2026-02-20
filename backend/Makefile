.PHONY: help build run test test-integration test-unit coverage wire mocks clean deps lint fmt migrate-up migrate-down migrate-create

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
