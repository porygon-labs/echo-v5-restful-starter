.PHONY: help build run dev test test-cover lint gosec fmt tidy clean up down migrate-up migrate-down migrate-status migrate-create migrate-redo module crud --crud

MODULE := $(shell awk '$$1 == "module" { print $$2; exit }' go.mod)
BIN   := ./bin/api

# ─── help ────────────────────────────────────────────────────────────────────

help: ## Show all targets
	@awk -F ':|##' '/^[a-zA-Z0-9_\-]+:.*##/ { printf "  \033[36m%-16s\033[0m %s\n", $$1, $$NF }' $(MAKEFILE_LIST)

# ─── build / run ─────────────────────────────────────────────────────────────

build: ## Build the binary
	go build -o $(BIN) ./cmd/api/main.go

run: ## Start the server
	go run ./cmd/api/main.go

dev: ## Start with hot-reload (requires air)
	air

# ─── test ────────────────────────────────────────────────────────────────────

test: ## Run unit tests
	go test -v -race ./...

test-cover: ## Run tests with coverage report
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo "\n── open HTML report ──────────────────────────────────────────"
	@echo "  go tool cover -html=coverage.out"

# ─── lint / security ────────────────────────────────────────────────────────

lint: ## Run golangci-lint
	golangci-lint run ./...

gosec: ## Run security scan
	gosec ./...

# ─── format ──────────────────────────────────────────────────────────────────

fmt: ## Sort imports and format code
	gci write -s standard -s "prefix($(MODULE))" -s default --custom-order \
		. 2>/dev/null; gofmt -w .

# ─── dependencies ────────────────────────────────────────────────────────────

tidy: ## Tidy module dependencies
	go mod tidy

# ─── clean ───────────────────────────────────────────────────────────────────

clean: ## Remove build artifacts
	rm -rf $(BIN) bin/ coverage.out

# ─── infra ───────────────────────────────────────────────────────────────────

up: ## Start dev dependencies (Postgres, Redis, Jaeger, Adminer)
	docker compose up -d

down: ## Stop dev dependencies
	docker compose down

# ─── migrations (goose) ─────────────────────────────────────────────────────

migrate-up: ## Run all pending migrations
	go run ./cmd/migrate/main.go up

migrate-down: ## Roll back the most recent migration
	go run ./cmd/migrate/main.go down

migrate-status: ## Print migration status
	go run ./cmd/migrate/main.go status

migrate-create: ## Create a new migration (e.g. make migrate-create name=add_users)
	go run ./cmd/migrate/main.go create $(name)

migrate-redo: ## Redo the last migration (down → up)
	go run ./cmd/migrate/main.go redo

# ─── generate ────────────────────────────────────────────────────────────────

# Usage:
#   make module name=book
#   make module name=book crud with=cache
#   make module name=book crud=true with=cache
#   make module name=book -- --crud --with=cache
#   make module name=book -- --crud --with=redis,migrations
# GNU Make requires `--` before custom long options.
module: ## Scaffold a new module (e.g. make module name=book -- --crud)
 	# make module name=sample -- --with=cache,migrations --crud
	@./scripts/create_module.sh "$(name)" \
		"$(if $(filter crud --crud,$(MAKECMDGOALS)),crud,crud=$(crud))" \
		"with=$(if $(with),$(with),$(--with))"

# Allows CRUD to be used as a flag-like make goal.
crud --crud:
	@:
