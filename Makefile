# =========================================================
# Project
# =========================================================
APP_NAME := petri-board-arena
GO := go

# =========================================================
# Paths
# =========================================================
API_MAIN := cmd/api/main.go
WORKER_MAIN := cmd/worker/main.go

MIGRATIONS_WRITE := migrations/write
MIGRATIONS_READ  := migrations/read

# =========================================================
# Env (local only)
# =========================================================
ifneq (,$(wildcard .env))
	include .env
	export
endif

# CQRS mapping (WRITE = Postgres, READ = Redis)
WRITE_DATABASE_URL ?= $(DATABASE_URL)
READ_DATABASE_URL  ?= $(REDIS_URL)

# =========================================================
# Tools
# =========================================================
BIN_DIR := bin
MIGRATE_BIN := $(BIN_DIR)/migrate
GQLGEN_BIN  := $(BIN_DIR)/gqlgen

.PHONY: tools tools-migrate tools-gqlgen
tools: tools-migrate tools-gqlgen

tools-migrate:
	@mkdir -p $(BIN_DIR)
	@if [ ! -x "$(MIGRATE_BIN)" ]; then \
		echo ">> installing golang-migrate into $(MIGRATE_BIN)"; \
		GOBIN=$(PWD)/bin go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	@$(MIGRATE_BIN) -version >/dev/null || true

tools-gqlgen:
	@mkdir -p $(BIN_DIR)
	@if [ ! -x "$(GQLGEN_BIN)" ]; then \
		echo ">> installing gqlgen into $(GQLGEN_BIN)"; \
		GOBIN=$$(pwd)/$(BIN_DIR) $(GO) install -mod=mod github.com/99designs/gqlgen@latest; \
	fi
	@$(GQLGEN_BIN) version >/dev/null || true

# =========================================================
# Help
# =========================================================
.PHONY: help
help:
	@echo ""
	@echo "Targets:"
	@echo "  dev                         Run API locally (loads .env if present)"
	@echo "  worker                      Run CQRS worker locally"
	@echo "  build                       Build API and Worker"
	@echo "  test                        Run all tests"
	@echo "  lint                        go vet + gofmt check"
	@echo "  gqlgen                      Generate GraphQL code"
	@echo ""
	@echo "Docker:"
	@echo "  up                          Start dependencies (docker compose)"
	@echo "  down                        Stop dependencies"
	@echo "  reset                       Stop dependencies and remove volumes"
	@echo ""
	@echo "Migrations (WRITE/Postgres):"
	@echo "  update                      Alias for migrate-write-up"
	@echo "  migrate-create-w name=...    Create write migration"
	@echo "  migrate-write-up             Apply write migrations"
	@echo "  migrate-write-down           Rollback last write migration"
	@echo "  migrate-write-version        Show current version"
	@echo ""
	@echo "Read-model (Redis):"
	@echo "  read-bootstrap               Init redis schema/version key (safe)"
	@echo "  read-flush                   Flush redis (DEV ONLY)"
	@echo ""
	@echo "CI:"
	@echo "  ci                          tools + gqlgen + lint + test"
	@echo "  ci-migrate-write-up          tools + migrate write up"
	@echo ""

# =========================================================
# Development
# =========================================================
.PHONY: dev
dev:
	@echo ">> running API"
	$(GO) run $(API_MAIN)

.PHONY: worker
worker:
	@echo ">> running CQRS worker"
	$(GO) run $(WORKER_MAIN)

# =========================================================
# Build
# =========================================================
.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	@echo ">> building API"
	$(GO) build -o $(BIN_DIR)/api $(API_MAIN)
	@echo ">> building Worker"
	$(GO) build -o $(BIN_DIR)/worker $(WORKER_MAIN)

# =========================================================
# Tests & Quality
# =========================================================
.PHONY: test
test:
	$(GO) test ./... -count=1

.PHONY: lint
lint:
	@echo ">> go vet"
	$(GO) vet ./...
	@echo ">> gofmt check"
	@test -z "$$(gofmt -l .)" || (echo ">> gofmt required. Run: gofmt -w ."; exit 1)

# =========================================================
# GraphQL
# =========================================================
.PHONY: gqlgen
gqlgen: tools-gqlgen
	@echo ">> gqlgen generate"
	$(GQLGEN_BIN) generate

# =========================================================
# Docker
# =========================================================
.PHONY: up
up:
	docker compose up -d

.PHONY: down
down:
	docker compose down

.PHONY: reset
reset:
	docker compose down -v

# =========================================================
# Migrations (WRITE/Postgres) - golang-migrate
# =========================================================
.PHONY: update migrate-create-w migrate-write-up migrate-write-down migrate-write-version

# make update == migrate up
update: migrate-write-up

migrate-create-w: tools-migrate
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create-w name=your_migration"; exit 1; fi
	@mkdir -p $(MIGRATIONS_WRITE)
	@echo ">> creating migration in $(MIGRATIONS_WRITE)"
	$(MIGRATE_BIN) create -ext sql -dir $(MIGRATIONS_WRITE) $(name)

migrate-write-up: tools-migrate
	@if [ -z "$(WRITE_DATABASE_URL)" ]; then echo "WRITE_DATABASE_URL/DATABASE_URL not set"; exit 1; fi
	@echo ">> migrate write up: $(MIGRATIONS_WRITE)"
	$(MIGRATE_BIN) -path $(MIGRATIONS_WRITE) -database "$(WRITE_DATABASE_URL)" up

migrate-write-down: tools-migrate
	@if [ -z "$(WRITE_DATABASE_URL)" ]; then echo "WRITE_DATABASE_URL/DATABASE_URL not set"; exit 1; fi
	@echo ">> migrate write down 1"
	$(MIGRATE_BIN) -path $(MIGRATIONS_WRITE) -database "$(WRITE_DATABASE_URL)" down 1

migrate-write-version: tools-migrate
	@if [ -z "$(WRITE_DATABASE_URL)" ]; then echo "WRITE_DATABASE_URL/DATABASE_URL not set"; exit 1; fi
	@echo ">> migrate write version"
	$(MIGRATE_BIN) -path $(MIGRATIONS_WRITE) -database "$(WRITE_DATABASE_URL)" version || true

# =========================================================
# Read-model (Redis) helpers
# =========================================================
.PHONY: read-bootstrap read-flush
read-bootstrap:
	@if [ -z "$(READ_DATABASE_URL)" ]; then echo "READ_DATABASE_URL/REDIS_URL not set"; exit 1; fi
	@echo ">> read bootstrap (redis): set schema version key"
	@redis-cli -u "$(READ_DATABASE_URL)" SETNX readmodel:schema_version 1 >/dev/null \
		&& echo "schema_version=1" || echo "schema_version already set"

read-flush:
	@if [ -z "$(READ_DATABASE_URL)" ]; then echo "READ_DATABASE_URL/REDIS_URL not set"; exit 1; fi
	@echo ">> FLUSHING REDIS (DEV ONLY)"
	@redis-cli -u "$(READ_DATABASE_URL)" FLUSHDB

# =========================================================
# CI targets (GitHub Actions)
# =========================================================
.PHONY: ci ci-migrate-write-up
ci: tools gqlgen lint test
	@echo ">> CI OK"

ci-migrate-write-up: tools-migrate
	@$(MAKE) migrate-write-up

# =========================================================
# Cleanup
# =========================================================
.PHONY: clean
clean:
	rm -rf $(BIN_DIR)