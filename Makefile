.PHONY: help build run test clean deps migrate dev lint fmt tidy

# Variables
APP_NAME = gin-boilerplate
MAIN_PATH = cmd/api/main.go
BUILD_DIR = bin
BINARY_NAME = $(BUILD_DIR)/$(APP_NAME)

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development targets
deps: ## Download dependencies
	go mod download
	go mod tidy

build: ## Build the application
	@mkdir -p $(BUILD_DIR)
	go build -o $(BINARY_NAME) $(MAIN_PATH)

run: ## Run the application
	go run $(MAIN_PATH)

dev: ## Run in development mode with hot reload (requires air)
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		exit 1; \
	fi

# Testing targets
test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Database targets
migrate-up: ## Run database migrations up
	@echo "Running database migrations..."
	# Add your migration command here

migrate-down: ## Run database migrations down
	@echo "Rolling back database migrations..."
	# Add your migration command here

# Quality targets
lint: ## Run linter
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

fmt: ## Format code
	go fmt ./...
	goimports -w .

tidy: ## Clean up dependencies
	go mod tidy

# Utility targets
clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	go clean -cache

install-tools: ## Install development tools
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/swaggo/swag/cmd/swag@latest

docs: ## Generate Swagger documentation
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/api/main.go -o docs; \
		echo "Swagger docs generated in docs/"; \
	else \
		echo "swag not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest"; \
		exit 1; \
	fi

swagger: docs ## Alias for docs command

# Docker targets
docker-build: ## Build Docker image
	docker build -t $(APP_NAME):latest .

docker-run: ## Run Docker container
	docker run -p 8080:8080 --env-file .env $(APP_NAME):latest

# Production targets
prod-build: ## Build for production
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BINARY_NAME) $(MAIN_PATH)

prod-run: ## Run production build
	./$(BINARY_NAME)

# Quick setup
setup: deps install-tools ## Quick setup for development
	@echo "Setup completed! You can now run 'make dev' to start development."