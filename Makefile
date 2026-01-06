.PHONY: help build run test clean migrate docker-up docker-down docker-build

help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building application..."
	@go build -o bin/docvault cmd/server/main.go

run: ## Run the application
	@echo "Running application..."
	@go run cmd/server/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/ dist/ build/ coverage.out coverage.html

migrate-up: ## Run database migrations up
	@echo "Running migrations up..."
	@go run cmd/server/main.go migrate up

migrate-down: ## Run database migrations down
	@echo "Running migrations down..."
	@go run cmd/server/main.go migrate down

migrate-create: ## Create a new migration file (usage: make migrate-create NAME=migration_name)
	@echo "Creating migration..."
	@migrate create -ext sql -dir internal/database/migrations -seq $(NAME)

docker-up: ## Start Docker containers
	@echo "Starting Docker containers..."
	@docker-compose up -d

docker-down: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	@docker-compose down

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t docvault-go:latest .

docker-logs: ## View Docker logs
	@docker-compose logs -f

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
