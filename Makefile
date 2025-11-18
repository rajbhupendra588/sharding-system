.PHONY: help build build-backend build-frontend run-router run-manager run-frontend start-backend start-frontend start-all test test-coverage lint fmt clean docker-build docker-up docker-down install-deps install-backend install-frontend

# Default target
.DEFAULT_GOAL := help

# Help target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Install dependencies
install-deps: install-backend install-frontend ## Install all dependencies

install-backend: ## Install Go dependencies
	@echo "Installing Go dependencies..."
	go mod download
	go mod tidy

install-frontend: ## Install frontend dependencies
	@echo "Installing frontend dependencies..."
	cd ui && npm install

# Build everything
build: build-backend build-frontend ## Build all components

build-backend: ## Build backend binaries
	@echo "Building backend..."
	@./scripts/build.sh

build-frontend: ## Build frontend
	@echo "Building frontend..."
	cd ui && npm run build

# Run individual services
run-router: ## Run router service
	@echo "Starting router service..."
	go run ./cmd/router

run-manager: ## Run manager service
	@echo "Starting manager service..."
	go run ./cmd/manager

run-frontend: ## Run frontend development server
	@echo "Starting frontend development server..."
	cd ui && npm run dev

# Start backend services (requires etcd)
start-backend: build-backend ## Start backend services
	@echo "Starting backend services..."
	@echo "Make sure etcd is running: docker-compose up -d etcd"
	@echo "Starting router on port 8080..."
	@./bin/router &
	@echo "Starting manager on port 8081..."
	@./bin/manager &
	@echo "Backend services started. Use 'make stop-backend' to stop them."

# Start frontend
start-frontend: install-frontend ## Start frontend
	@echo "Starting frontend UI on port 3000..."
	cd ui && npm run dev

# Start everything (backend + frontend)
start-all: install-deps ## Start all services
	@echo "Starting all services..."
	@echo "1. Starting etcd..."
	@docker-compose up -d etcd
	@sleep 3
	@echo "2. Building backend..."
	@$(MAKE) build-backend
	@echo "3. Starting backend services..."
	@$(MAKE) start-backend
	@sleep 2
	@echo "4. Starting frontend..."
	@$(MAKE) start-frontend

# Testing
test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@./scripts/test.sh

# Code quality
lint: ## Run linters
	@./scripts/lint.sh

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	else \
		echo "goimports not found, skipping import formatting"; \
	fi

# Cleanup
clean: ## Clean build artifacts
	@./scripts/clean.sh

# Docker operations
docker-build: ## Build Docker images
	@echo "Building Docker images..."
	docker build -f Dockerfile.router -t sharding-router .
	docker build -f Dockerfile.manager -t sharding-manager .

docker-up: ## Start Docker containers
	@echo "Starting Docker containers..."
	docker-compose up -d

docker-down: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f

