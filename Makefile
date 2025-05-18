# Makefile for TradeMicro Go API

# Load environment variables from .env.deploy if present
ifneq (,$(wildcard .env.deploy))
	include .env.deploy
	export
endif

APP_NAME=trademicro
MIGRATIONS_DIR=./migrations
CSV_PROCESSOR=./csv_to_db.py
CSV_DIR=./instrument_data
CSV_SOURCE_URL?=https://api.dhan.co/v2/instrument/NSE_EQ

.PHONY: all build run migrate clean build-migrate migrate-up migrate-down migrate-version migrate-create migrate-steps-up migrate-steps-down migrate-force migrate-reset import-csv import-csv-all import-csv-compact import-csv-detailed check-csv

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

# Import single CSV file to database
# Usage: make import-csv file=path/to/file.csv mode=compact
import-csv:
	@if [ -z "$(file)" ]; then \
		echo "Error: file path is required. Usage: make import-csv file=path/to/file.csv mode=compact"; \
		exit 1; \
	fi
	@echo "Importing CSV file: $(file)"
	@python $(CSV_PROCESSOR) --file $(file) --mode $(mode) --batch-size 1000

# Import all CSV files from instrument_data directory
# This processes all batch files in order
import-csv-all:
	@echo "Importing all CSV files from $(CSV_DIR)..."
	@for file in $(CSV_DIR)/instruments_batch_*.csv; do \
		echo "Processing $$file"; \
		python $(CSV_PROCESSOR) --file $$file --mode detailed --batch-size 1000; \
	done

# Import specifically the compact format CSV
import-csv-compact:
	@echo "Importing compact format CSV..."
	@if [ -f "dhan_instruments.csv" ]; then \
		python $(CSV_PROCESSOR) --file dhan_instruments.csv --mode compact --batch-size 1000; \
	else \
		echo "Error: dhan_instruments.csv not found"; \
		exit 1; \
	fi

# Import specifically the detailed format CSVs
import-csv-detailed:
	@echo "Importing detailed format CSVs..."
	@if [ -d "$(CSV_DIR)" ]; then \
		python $(CSV_PROCESSOR) --file "$(CSV_DIR)/instruments_batch_001.csv" --mode detailed --batch-size 1000; \
	else \
		echo "Error: $(CSV_DIR) directory not found"; \
		exit 1; \
	fi

# Check available CSV files
check-csv:
	@echo "Checking available CSV files:"
	@echo "Main CSV file:"
	@ls -la dhan_instruments.csv 2>/dev/null || echo "dhan_instruments.csv not found"
	@echo "\nBatch CSV files:"
	@if [ -d "$(CSV_DIR)" ]; then \
		ls -la $(CSV_DIR)/instruments_batch_*.csv 2>/dev/null | head -10; \
		count=$$(ls -1 $(CSV_DIR)/instruments_batch_*.csv 2>/dev/null | wc -l); \
		if [ $$count -gt 10 ]; then \
			echo "... and $$((count-10)) more files"; \
		fi; \
	else \
		echo "$(CSV_DIR) directory not found"; \
	fi

# Update and import CSV data in one step
# This downloads the latest CSV and imports it
update-csv-import: 
	@echo "Downloading latest instrument data and importing..."
	@if curl -s -H "Accept: text/csv" -o dhan_instruments_latest.csv $(CSV_SOURCE_URL); then \
		echo "Download successful. Checking file..."; \
		if [ -s dhan_instruments_latest.csv ]; then \
			head -n 1 dhan_instruments_latest.csv; \
			echo "Importing with upsert mode..."; \
			python $(CSV_PROCESSOR) --file dhan_instruments_latest.csv --mode compact --upsert; \
			echo "Import completed."; \
		else \
			echo "Error: Downloaded file is empty"; \
			exit 1; \
		fi; \
	else \
		echo "Error: Failed to download CSV file"; \
		exit 1; \
	fi

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
	@echo "  make import-csv file=path/to/file.csv mode=compact - Import single CSV file to database"
	@echo "  make import-csv-all          - Import all CSV files from instrument_data directory"
	@echo "  make import-csv-compact      - Import specifically the compact format CSV"
	@echo "  make import-csv-detailed     - Import specifically the detailed format CSVs"
	@echo "  make check-csv               - Check available CSV files"
	@echo "  make update-csv-import       - Update and import CSV data in one step"
