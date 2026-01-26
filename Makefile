.PHONY: help run build dev test clean db-up db-down db-migrate lint format

help:
	@echo "Available commands:"
	@echo "  make run          - Run the server"
	@echo "  make build        - Build the server binary"
	@echo "  make dev          - Run server in development mode (requires air)"
	@echo "  make test         - Run all tests"
	@echo "  make test-v       - Run tests with verbose output"
	@echo "  make lint         - Run linter (requires golangci-lint)"
	@echo "  make format       - Format code with gofmt"
	@echo "  make vet          - Run go vet"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make db-up        - Start database with Docker Compose"
	@echo "  make db-down      - Stop database with Docker Compose"
	@echo "  make db-migrate   - Run database migrations (requires sqlc)"
	@echo "  make sqlc         - Generate Go code from SQL queries"

# Build and run commands
run:
	@echo "Starting server..."
	go run ./cmd/api

build:
	@echo "Building server..."
	go build -o bin/formify-server ./cmd/api

dev:
	@echo "Starting server in development mode..."
	air

# Testing commands
test:
	go test ./...

test-v:
	go test -v ./...

# Code quality commands
lint:
	golangci-lint run ./...

format:
	gofmt -s -w .

vet:
	go vet ./...

# Cleanup
clean:
	rm -rf bin/
	go clean

# Database commands
db-up:
	@echo "Starting database..."
	docker compose -f infrastructure/docker-compose.yml up -d db

db-down:
	@echo "Stopping database..."
	docker compose -f infrastructure/docker-compose.yml down

db-migrate:
	@echo "Running migrations..."
	cd internal/database && sqlc generate

sqlc:
	@echo "Generating code from SQL..."
	cd internal/database && sqlc generate

# Installation helpers
install-dev-tools:
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
