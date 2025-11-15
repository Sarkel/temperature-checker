# Makefile for project-donkey-backend-go
# This Makefile helps with installing and operating SQLC and go-migration tools

# Variables
SQLC_VERSION := latest
MIGRATE_VERSION := latest
MIGRATIONS_DIR := migrations

# Include .env file if it exists
-include .env.defaults
-include .env

# Set default DB_URL if not defined in .env
DB_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)


.PHONY: install-tools sqlc migrate-up migrate-down migrate-create help env-setup

# Default target
help:
	@echo "Available targets:"
	@echo "  install-tools    - Install SQLC and golang-migrate tools"
	@echo "  sqlc             - Generate Go code from SQL queries using SQLC"
	@echo "  migrate-up       - Run all pending migrations"
	@echo "  migrate-down     - Rollback the last migration"
	@echo "  migrate-create   - Create a new migration file (usage: make migrate-create name=migration_name)"
	@echo "  env-setup        - Create a .env file from the template (won't overwrite existing .env)"
	@echo ""
	@echo "Environment variables:"
	@echo "  Database connection settings can be configured in a .env file"
	@echo "  Copy .env.template to .env and modify as needed"

# Install required tools
install-tools:
	@echo "Installing SQLC and golang-migrate..."
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)
	@echo "Tools installed successfully!"

# Generate Go code from SQL queries using SQLC
sqlc:
	@echo "Generating Go code from SQL queries..."
	sqlc generate
	@echo "Code generation completed!"

# Run all pending migrations
migrate-up:
	@echo "Running migrations..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up
	@echo "Migrations completed!"

# Rollback the last migration
migrate-down:
	@echo "Rolling back the last migration..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1
	@echo "Rollback completed!"

# Create a new migration file
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: Migration name not provided. Usage: make migrate-create name=migration_name"; \
		exit 1; \
	fi
	@echo "Creating new migration: $(name)"
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)
	@echo "Migration file created!"

# Create a .env file from the template
env-setup:
	@if [ -f .env ]; then \
		echo ".env file already exists. Remove it first if you want to create a new one."; \
	else \
		cp .env.defaults .env; \
		echo ".env file created from template. Please update it with your credentials."; \
	fi