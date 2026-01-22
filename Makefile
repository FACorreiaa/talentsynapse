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
DB_DSN ?= postgres://postgres:postgres@localhost:5470/$(PROJECT_NAME)?sslmode=disable
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

dev: ## Start development server with live reload (using Air)
	@echo "🚀 Starting development server with Air..."
	@GO_ENV=development air

# run templ generation in watch mode to detect all .templ files and 
# re-create _templ.txt files on change, then send reload event to browser. 
# Default url: http://localhost:7331
live/templ:
	templ generate --watch --proxy="http://localhost:8081" --open-browser=false -v

# run air to detect any go file changes to re-build and re-run the server.
live/server:
	go run github.com/air-verse/air@v1.63.0 \
	--build.cmd "go build -o tmp/bin/main ./cmd/server" --build.bin "tmp/bin/main" --build.delay "100" \
	--build.include_ext "go" \
	--build.stop_on_error "false" \
	--misc.clean_on_exit true

# run tailwindcss to generate the styles.css bundle in watch mode.
live/tailwind:
	tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --minify --watch

# watch for any js or css change in the assets/ folder, then reload the browser via templ proxy.
live/sync_assets:
	go run github.com/air-verse/air@v1.63.0 \
	--build.cmd "templ generate --notify-proxy" \
	--build.bin "/usr/bin/true" \
	--build.delay "100" \
	--build.exclude_dir "tmp,vendor,bin" \
	--build.include_dir "assets" \
	--build.include_ext "js,css"

# start all 5 watch processes in parallel.
live: 
	make -j5 live/templ live/server live/tailwind live/sync_assets

# =========================================================================
# Build
# =========================================================================

build: templ ## Build production binary
	@echo "🔨 Building CSS..."
	@tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --minify
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
# Git Hooks
# =========================================================================

hooks-install: ## Install Git hooks (pre-commit)
	@echo "📦 Installing Git hooks..."
	@git config core.hooksPath .githooks
	@chmod +x .githooks/*
	@echo "✅ Git hooks installed! Commits will now run: templ fmt, gofumpt, golangci-lint"

hooks-uninstall: ## Uninstall Git hooks
	@git config --unset core.hooksPath
	@echo "✅ Git hooks uninstalled"

pre-commit: ## Run pre-commit checks manually
	@templ fmt .
	@gofumpt -l -w .
	@templ generate
	@golangci-lint run --fix

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
	templ generate --watch --proxy="http://localhost:8081" --cmd="go run ."



t-fmt:
	templ fmt .

# Manual Tailwind CSS commands (not needed if using 'make dev')
tailwind:
	tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --watch

tailwind-build:
	tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --minify

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
