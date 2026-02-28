# Makefile for Vigilate

# === Configurable variables ===
DB_USER ?= postgres
DB_PASS ?= secret
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_NAME ?= vigilate

APP_PORT ?= :4000
IDENTIFIER ?= vigilate
DOMAIN ?= localhost
IN_PRODUCTION ?= false

# Mail and pusher defaults
PUSHER_HOST ?= localhost
PUSHER_PORT ?= 4001
PUSHER_APP ?= 1
PUSHER_KEY ?= abc123
PUSHER_SECRET ?= 123abc
PUSHER_SECURE ?= false

# === Commands ===
GO_RUN := go run ./cmd/web

# === Targets ===

# Run the app locally with DB and flags
run:
	$(GO_RUN) \
		-dbuser=$(DB_USER) \
		-dbpass=$(DB_PASS) \
		-dbhost=$(DB_HOST) \
		-dbport=$(DB_PORT) \
		-db=$(DB_NAME) \
		-port=$(APP_PORT) \
		-identifier=$(IDENTIFIER) \
		-domain=$(DOMAIN) \
		-production=$(IN_PRODUCTION) \
		-pusherHost=$(PUSHER_HOST) \
		-pusherPort=$(PUSHER_PORT) \
		-pusherApp=$(PUSHER_APP) \
		-pusherKey=$(PUSHER_KEY) \
		-pusherSecret=$(PUSHER_SECRET) \
		-pusherSecure=$(PUSHER_SECURE)

# Run migrations using soda
migrate-up:
	soda migrate up

migrate-down:
	soda migrate down

# Reset DB: roll back all migrations, then apply
migrate-reset: migrate-down migrate-up

# Build binary for production
build:
	go build -o bin/vigilate ./cmd/web

# Clean build artifacts
clean:
	rm -rf bin/*

# Show environment
env:
	@echo "DB_USER=$(DB_USER)"
	@echo "DB_PASS=$(DB_PASS)"
	@echo "DB_HOST=$(DB_HOST)"
	@echo "DB_PORT=$(DB_PORT)"
	@echo "DB_NAME=$(DB_NAME)"
	@echo "APP_PORT=$(APP_PORT)"
	@echo "IDENTIFIER=$(IDENTIFIER)"
	@echo "DOMAIN=$(DOMAIN)"