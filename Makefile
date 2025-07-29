# Simple Makefile for a Go project

# Build the application
all: build test

build:
	@echo "Building..."
	@go build -o main cmd/main.go

# Run the application
run:
	@go run cmd/main.go
# Create DB container
docker-run:
	docker compose up -d --build --remove-orphans

# Shutdown DB container
docker-down:
	docker compose down

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v
# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

# Generate SQLC
sqlc:
	@sqlc generate --file ./internal/db/sqlc.yaml

# Usage: make migrate-create name=your_migration_name
migrate-create:
	goose create $(name) sql

migrate-up:
	@goose up sql

migrate-down:
	@goose down sql
# Run migrations
#migrate-up:
#	@set -a && source .env && set +a && goose -dir ./internal/infrastructure/persistence/sql postgres "user=$${DB_USERNAME} password=$${DB_PASSWORD} host=localhost port=$${DB_PORT} dbname=$${DB_DATABASE} sslmode=disable" up

# Rollback migrations
# migrate-down:
# 	@set -a && source .env && set +a && goose -dir ./internal/infrastructure/persistence/sql postgres "user=$${DB_USERNAME} password=$${DB_PASSWORD} host=localhost port=$${DB_PORT} dbname=$${DB_DATABASE} sslmode=disable" down

.PHONY: all build run test clean watch docker-run docker-down itest