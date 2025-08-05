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

#migrate-up:
#	@goose up sql

#migrate-down: @goose down sql
# Run migrations
#migrate-up:
#	@set -a && source .env && set +a && goose -dir ./internal/infrastructure/persistence/sql postgres "user=$${DB_USERNAME} password=$${DB_PASSWORD} host=localhost port=$${DB_PORT} dbname=$${DB_DATABASE} sslmode=disable" up

# Rollback migrations
# migrate-down:
# 	@set -a && source .env && set +a && goose -dir ./internal/infrastructure/persistence/sql postgres "user=$${DB_USERNAME} password=$${DB_PASSWORD} host=localhost port=$${DB_PORT} dbname=$${DB_DATABASE} sslmode=disable" down


# Applies all pending migrations inside the running container.
migrate-up:
	@echo "Running migrations up inside the Docker container..."
	@docker-compose exec trip-plan-service goose -dir "./internal/db/migrations" up

# Rolls back the single most recent migration inside the running container.
migrate-down:
	@echo "Running one migration down inside the Docker container..."
	@docker-compose exec trip-plan-service goose -dir "./internal/db/migrations" down

# Checks the status of migrations inside the running container.
migrate-status:
	@echo "Checking migration status inside the Docker container..."
	@docker-compose exec trip-plan-service goose -dir "./internal/db/migrations" status


db-reset:
	@echo "\033[0;31mDANGER: This command will completely destroy the database and all its data.\033[0m"
	@read -p "Are you absolutely sure you want to continue? [y/N] " choice; \
	if [ "$$choice" = "y" ] || [ "$$choice" = "Y" ]; then \
		echo "Stopping and removing containers, volumes, and networks..."; \
		docker-compose down -v --remove-orphans; \
		echo "\033[0;32mServices and volumes destroyed.\033[0m"; \
		echo "Restarting services (this will create a new, empty database)..."; \
		make docker-run; \
		echo "Waiting for the database to become healthy..."; \
		sleep 8; \
		echo "Applying all migrations from scratch..."; \
		make migrate-up; \
		echo "\033[0;32mDatabase reset complete. All migrations applied.\033[0m"; \
	else \
		echo "Aborted by user."; \
		exit 1; \
	fi



.PHONY: all build run test clean watch docker-run docker-down itest