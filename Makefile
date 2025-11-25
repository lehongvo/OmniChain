.PHONY: help build test clean docker-build docker-up docker-down migrate-up migrate-down lint security-scan docker-build-service

# Variables
DOCKER_REGISTRY ?= onichange
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
SERVICES = api-gateway order-service user-service store-service payment-service inventory-service notification-service

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all services
	@echo "Building all services..."
	@for service in $(SERVICES); do \
		echo "Building $$service..."; \
		go build -ldflags="-w -s" -o bin/$$service ./cmd/$$service || exit 1; \
	done
	@echo "Build complete!"

build-service: ## Build a specific service (usage: make build-service SERVICE=api-gateway)
	@if [ -z "$(SERVICE)" ]; then \
		echo "Error: SERVICE is required. Usage: make build-service SERVICE=api-gateway"; \
		exit 1; \
	fi
	@echo "Building $(SERVICE)..."
	@go build -ldflags="-w -s" -o bin/$(SERVICE) ./cmd/$(SERVICE)

test: ## Run all tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

test-unit: ## Run unit tests only
	@go test -v -short ./...

test-integration: ## Run integration tests
	@go test -v -tags=integration ./tests/integration/...

lint: ## Run linters
	@echo "Running linters..."
	@golangci-lint run ./...

security-scan: ## Run security scans
	@echo "Running security scans..."
	@gosec ./...
	@govulncheck ./...

migrate-up: ## Run database migrations up
	@migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" up

migrate-down: ## Rollback database migrations
	@migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" down

docker-build: ## Build all Docker images
	@echo "Building Docker images..."
	@for service in $(SERVICES); do \
		echo "Building $$service..."; \
		docker build \
			--build-arg SERVICE=$$service \
			--build-arg BUILD_REF=$(VERSION) \
			--build-arg BUILD_DATE=$(BUILD_DATE) \
			-f deployments/docker/Dockerfile \
			-t $(DOCKER_REGISTRY)/$$service:$(VERSION) \
			-t $(DOCKER_REGISTRY)/$$service:latest \
			. || exit 1; \
	done
	@echo "All Docker images built successfully!"

docker-build-service: ## Build a specific service Docker image (usage: make docker-build-service SERVICE=api-gateway)
	@if [ -z "$(SERVICE)" ]; then \
		echo "Error: SERVICE is required. Usage: make docker-build-service SERVICE=api-gateway"; \
		exit 1; \
	fi
	@echo "Building Docker image for $(SERVICE)..."
	@docker build \
		--build-arg SERVICE=$(SERVICE) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-f deployments/docker/Dockerfile \
		-t $(DOCKER_REGISTRY)/$(SERVICE):$(VERSION) \
		-t $(DOCKER_REGISTRY)/$(SERVICE):latest \
		.

docker-up: ## Start all services with Docker Compose
	@cd deployments/docker && docker-compose up -d

docker-down: ## Stop all Docker services
	@cd deployments/docker && docker-compose down

docker-logs: ## View Docker logs (usage: make docker-logs SERVICE=api-gateway)
	@if [ -z "$(SERVICE)" ]; then \
		cd deployments/docker && docker-compose logs -f; \
	else \
		cd deployments/docker && docker-compose logs -f $(SERVICE); \
	fi

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf coverage.out
	@go clean -cache

deps: ## Download dependencies
	@go mod download
	@go mod tidy

.PHONY: proto
proto: ## Generate gRPC code from proto files
	@echo "Generating gRPC code..."
	@for proto_dir in proto/*/; do \
		service=$$(basename $$proto_dir); \
		echo "Generating code for $$service..."; \
		protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			$$proto_dir/*.proto || exit 1; \
	done
	@echo "gRPC code generation complete!"
