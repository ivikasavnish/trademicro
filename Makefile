# Makefile for TradeMicro Go API

# Load environment variables from .env.deploy if present
ifneq (,$(wildcard .env.deploy))
	include .env.deploy
	export
endif

APP_NAME=trademicro
MIGRATIONS_DIR=./migrations

.PHONY: all build run migrate clean

all: build

build:
	go build -o $(APP_NAME) .

run:
	PORT=$(PORT) REDIS_URL=$(REDIS_URL) SECRET_KEY=$(SECRET_KEY) POSTGRES_URL=$(POSTGRES_URL) ./$(APP_NAME)

migrate:
	migrate -database "$(POSTGRES_URL)" -path $(MIGRATIONS_DIR) up

clean:
	rm -f $(APP_NAME)
