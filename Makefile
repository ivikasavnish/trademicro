# Makefile for TradeMicro Go API

# Load environment variables from .env.deploy if present
ifneq (,$(wildcard .env.deploy))
	include .env.deploy
	export
endif

APP_NAME=trademicro
MIGRATIONS_DIR=./migrations

.PHONY: all build run migrate clean build-migrate migrate-up migrate-down migrate-version migrate-create migrate-steps-up migrate-steps-down migrate-force migrate-reset

# Database connection string (update as needed)
DB_URL ?= "postgresql://postgres:password@localhost:5432/trademicro?sslmode=disable"

all: build

build:
	go build -o $(APP_NAME) .

build-migrate:
	@echo "Building migration tool..."
	@go build -o bin/migrate cmd/migrate/main.go
	@echo "Done. Binary created at bin/migrate"

# Run migrations up
migrate-up:
	@migrate -database $(DB_URL) -path ./migrations up

# Run migrations down
migrate-down:
	@migrate -database $(DB_URL) -path ./migrations down

# Show current migration version
migrate-version:
	@migrate -database $(DB_URL) -path ./migrations version

# Create a new migration
# Usage: make migrate-create name=create_users_table
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: name is required. Usage: make migrate-create name=create_users_table"; \
		exit 1; \
	fi
	@go run cmd/migrate/main.go -create=$(name)

# Run a specific number of migrations up
# Usage: make migrate-steps-up steps=1
migrate-steps-up:
	@if [ -z "$(steps)" ]; then \
		echo "Error: steps is required. Usage: make migrate-steps-up steps=1"; \
		exit 1; \
	fi
	@migrate -database $(DB_URL) -path ./migrations up $(steps)

# Run a specific number of migrations down
# Usage: make migrate-steps-down steps=1
migrate-steps-down:
	@if [ -z "$(steps)" ]; then \
		echo "Error: steps is required. Usage: make migrate-steps-down steps=1"; \
		exit 1; \
	fi
	@migrate -database $(DB_URL) -path ./migrations down $(steps)

# Force migration version (use with caution)
# Usage: make migrate-force version=1
migrate-force:
	@if [ -z "$(version)" ]; then \
		echo "Error: version is required. Usage: make migrate-force version=1"; \
		exit 1; \
	fi
	@migrate -database $(DB_URL) -path ./migrations force $(version)

# Reset all migrations (drops and recreates everything)
migrate-reset:
	@echo "WARNING: This will reset ALL migrations. Press Ctrl+C to cancel..."
	@sleep 5
	@migrate -database $(DB_URL) -path ./migrations force 0
	@migrate -database $(DB_URL) -path ./migrations up

# Install required packages
install-deps:
	@go get -u github.com/golang-migrate/migrate/v4
	@go get -u github.com/golang-migrate/migrate/v4/database/postgres
	@go get -u github.com/golang-migrate/migrate/v4/source/file
	@go get -u github.com/joho/godotenv
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

run:
	PORT=$(PORT) REDIS_URL=$(REDIS_URL) SECRET_KEY=$(SECRET_KEY) POSTGRES_URL=$(POSTGRES_URL) ./$(APP_NAME)

migrate:
	migrate -database "$(POSTGRES_URL)" -path $(MIGRATIONS_DIR) up

clean:
	rm -f $(APP_NAME)

# Help command
help:
	@echo "Available commands:"
	@echo "  make build-migrate           - Build the migration tool"
	@echo "  make migrate-up              - Run all migrations up"
	@echo "  make migrate-down            - Run all migrations down"
	@echo "  make migrate-version         - Show current migration version"
	@echo "  make migrate-create name=X   - Create a new migration named X"
	@echo "  make migrate-steps-up steps=N - Run N migrations up"
	@echo "  make migrate-steps-down steps=N - Run N migrations down"
	@echo "  make migrate-force version=N - Force migration version to N"
	@echo "  make migrate-reset           - Reset and rerun all migrations"
	@echo "  make install-deps            - Install required dependencies"
