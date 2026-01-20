# GoForge Project Makefile
# =========================================================================
# Usage:
#   make setup    - Install all tools (Air, Templ, Goose, Tailwind + DaisyUI)
#   make dev      - Start development server with live reload
#   make build    - Build production binary
#   make test     - Run tests
# =========================================================================

PROJECT_NAME := myapp
BINARY_NAME := server

# Database settings (override via .env or environment)
DB_DSN ?= postgres://localhost:5432/$(PROJECT_NAME)?sslmode=disable
GOOSE_DRIVER := postgres
GOOSE_MIGRATION_DIR := ./internal/database/migrations

.PHONY: all build run test clean dev setup help

all: build

# =========================================================================
# Setup & Dependencies
# =========================================================================

setup: ## Install all development tools
	@echo "📦 Installing Go tools..."
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/air-verse/air@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	@echo ""
	@echo "⚠️  Note: Using standalone Tailwind CLI (must be installed separately)"
	@echo "   macOS: brew install tailwindcss"
	@echo "   Linux: Download from https://github.com/tailwindlabs/tailwindcss/releases"
	@echo ""
	@echo "✅ Setup complete! Run 'make dev' to start development."

tidy: ## Tidy Go modules
	go mod tidy

ci-setup: ## Setup for CI environments (GitHub Actions)
	@echo "📦 Installing Go tools for CI..."
	go install github.com/a-h/templ/cmd/templ@latest
	@echo "📥 Installing Tailwind CSS standalone binary..."
	@echo "See: https://tailwindcss.com/blog/standalone-cli"
	@echo "✅ CI setup complete!"

# =========================================================================
# Development
# =========================================================================

dev: ## Start development server with live reload (Air handles Templ + Tailwind + Go)
	@echo "🚀 Starting development server with Air..."
	@echo "📝 Watching: .templ → template generation → Tailwind CSS rebuild → Go binary"
	@GO_ENV=development air

templ: ## Generate Templ templates
	templ generate

css: ## Build CSS once
	tailwindcss -i ./assets/css/index.css -o ./assets/css/output.css --minify

# =========================================================================
# Build
# =========================================================================

build: templ ## Build production binary
	@echo "🔨 Building CSS..."
	@tailwindcss -i ./assets/css/index.css -o ./assets/css/output.css --minify
	@echo "🔨 Building binary..."
	@CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/$(BINARY_NAME) ./cmd/server
	@echo "✅ Build complete: ./bin/$(BINARY_NAME)"

run: build ## Build and run the application
	./bin/$(BINARY_NAME)

clean: ## Remove build artifacts
	rm -rf bin/
	rm -rf tmp/
	rm -f assets/css/output.css
	find . -name "*_templ.go" -delete

# =========================================================================
# Testing
# =========================================================================

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run golangci-lint
	golangci-lint run

# =========================================================================
# Database (Goose)
# =========================================================================

db-status: ## Show migration status
	goose -dir $(GOOSE_MIGRATION_DIR) $(GOOSE_DRIVER) "$(DB_DSN)" status

db-up: ## Run all pending migrations
	goose -dir $(GOOSE_MIGRATION_DIR) $(GOOSE_DRIVER) "$(DB_DSN)" up

db-down: ## Rollback the last migration
	goose -dir $(GOOSE_MIGRATION_DIR) $(GOOSE_DRIVER) "$(DB_DSN)" down

db-reset: ## Rollback all migrations
	goose -dir $(GOOSE_MIGRATION_DIR) $(GOOSE_DRIVER) "$(DB_DSN)" reset

db-create: ## Create a new migration (usage: make db-create name=create_users)
	@read -p "Enter migration name: " name; \
	goose -dir $(GOOSE_MIGRATION_DIR) create $$name sql

# =========================================================================
# Docker
# =========================================================================

docker-build: ## Build Docker image
	docker build -t $(PROJECT_NAME) .

docker-run: ## Run Docker container
	docker run -p 8080:8080 --env-file .env $(PROJECT_NAME)

docker-compose-up: ## Start with docker-compose
	docker-compose up -d

docker-compose-down: ## Stop docker-compose
	docker-compose down

# =========================================================================
# Help
# =========================================================================

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Run templ generation in watch mode
templ:
	templ fmt . && templ generate --watch --proxy="http://localhost:7000" --open-browser=false

t-fmt:
	templ fmt .

# Manual Tailwind CSS commands (not needed if using 'make dev')
tailwind:
	tailwindcss -i ./assets/css/index.css -o ./assets/css/output.css --watch

tailwind-build:
	tailwindcss -i ./assets/css/index.css -o ./assets/css/output.css --minify

# =========================================================================
# Build Analysis
# =========================================================================

size-report: ## Show size breakdown of build artifacts
	@echo "📊 Build Size Report"
	@echo "===================="
	@echo ""
	@echo "Binary:"
	@ls -lh bin/server 2>/dev/null | awk '{print "  Size:", $$5}' || echo "  Not built (run 'make build' first)"
	@echo ""
	@echo "CSS:"
	@ls -lh assets/css/output.css 2>/dev/null | awk '{print "  Size:", $$5}' || echo "  Not built"
	@echo ""
	@echo "Total Assets:"
	@du -sh assets/ 2>/dev/null | awk '{print "  Size:", $$1}' || echo "  Not found"
	@echo ""
	@echo "Binary Contents:"
	@du -sh bin/ 2>/dev/null | awk '{print "  Total:", $$1}' || echo "  Not built"

build-prod: ## Build production binary with optimization report
	@./scripts/build-prod.sh
