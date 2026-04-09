ifneq ($(wildcard .env),)
-include .env
export
else
$(warning WARNING: .env file not found! Using .env.example)
include .env.example
export
endif

BASE_STACK = docker compose -f docker-compose.yml
INTEGRATION_TEST_STACK = $(BASE_STACK) -f docker-compose-integration-test.yml
ALL_STACK = $(INTEGRATION_TEST_STACK)

# HELP =================================================================================================================
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help

help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

compose-up: ### Run docker compose (without backend and reverse proxy)
	$(BASE_STACK) up --build -d db && docker compose logs -f
.PHONY: compose-up

compose-up-all: ### Run docker compose (with backend and reverse proxy)
	$(BASE_STACK) up --build -d
.PHONY: compose-up-all

compose-up-integration-test: ### Run docker compose with integration test
	$(INTEGRATION_TEST_STACK) up --build --abort-on-container-exit --exit-code-from integration-test
.PHONY: compose-up-integration-test

compose-down: ### Down docker compose
	$(ALL_STACK) down --remove-orphans
.PHONY: compose-down

swag: ### generate swagger documentation
	swag init --parseDependency --parseInternal -g cmd/app/main.go -o docs/swagger
.PHONY: swag

deps: ### deps tidy + verify
	go mod tidy && go mod verify
.PHONY: deps

deps-audit: ### check dependencies vulnerabilities
	govulncheck ./...
.PHONY: deps-audit

format: ### Run code formatter
	gofumpt -l -w .
	gci write . --skip-generated -s standard -s default
.PHONY: format

kill-port: ### kill process running on HTTP_PORT
	@PORT=$${HTTP_PORT:-8080}; \
	PID=$$(lsof -ti tcp:$$PORT 2>/dev/null); \
	if [ -n "$$PID" ]; then \
		echo "Port $$PORT da ishlaayotgan process (PID: $$PID) o'chirilmoqda..."; \
		kill -9 $$PID; \
		sleep 0.5; \
	fi
.PHONY: kill-port

run: kill-port ### run application with all code generation
	swag init --parseDependency --parseInternal -g cmd/app/main.go -o docs/swagger > /dev/null 2>&1 && \
	CGO_ENABLED=0 go run ./cmd/app
.PHONY: run

docker-rm-volume: ### remove docker volume
	docker volume rm go-clean-template_pg-data
.PHONY: docker-rm-volume

linter-golangci: ### check by golangci linter
	golangci-lint run
.PHONY: linter-golangci

arch-check: ### verify bounded context & DDD architecture rules
	go test -v -count=1 ./test/arch/...
.PHONY: arch-check

linter-hadolint: ### check by hadolint linter
	git ls-files --exclude='Dockerfile*' --ignored | xargs hadolint
.PHONY: linter-hadolint

linter-dotenv: ### check by dotenv linter
	dotenv-linter
.PHONY: linter-dotenv

test: ### run test
	go test -v -race -covermode atomic -coverprofile=coverage.txt ./internal/... ./pkg/...
.PHONY: test

test-fuzz: ### run fuzz tests
	go test -fuzz=FuzzGetPasswordStrength -fuzztime=30s ./pkg/validation
.PHONY: test-fuzz

test-prop: ### run property-based tests
	go test -v -run TestSanitizePhone_Property ./pkg/validation
.PHONY: test-prop

integration-test: ### run integration-test
	go clean -testcache && go test -v ./integration-test/...
.PHONY: integration-test

mock: ### run mockgen
	mockgen -source ./internal/repo/contracts.go -package usecase_test > ./internal/usecase/mocks_repo_test.go
	mockgen -source ./internal/usecase/contracts.go -package usecase_test > ./internal/usecase/mocks_usecase_test.go
.PHONY: mock

# ==============================================================================
# Goose Migration Targets
# ==============================================================================

CURRENT_DIR = $(shell pwd)
export GOOSE_MIGRATION_DIR = $(CURRENT_DIR)/migrations/postgres
export GOOSE_DRIVER = postgres
export GOOSE_DBSTRING = "host=$(POSTGRES_HOST) port=$(POSTGRES_PORT) dbname=$(POSTGRES_DB) user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) sslmode=$(POSTGRES_SSLMODE)"
GOOSE_ENV = GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING)

migration-create: ### create new migration (interactive)
	@read -p "Enter migration name: " MIGRATION_NAME; \
	echo $$MIGRATION_NAME; \
	goose -s create $$MIGRATION_NAME sql
.PHONY: migration-create

migration-up: ### run all available migrations
	$(GOOSE_ENV) goose up -v
.PHONY: migration-up

migration-down: ### roll back the version by 1
	$(GOOSE_ENV) goose down -v
.PHONY: migration-down

migration-status: ### dump the migration status for the current DB
	$(GOOSE_ENV) goose status
.PHONY: migration-status

migration-redo: ### re-run the latest migration
	$(GOOSE_ENV) goose redo -v
.PHONY: migration-redo

migration-reset: ### roll back all migrations
	$(GOOSE_ENV) goose reset -v
.PHONY: migration-reset

migration-validate: ### check migration files without running them
	$(GOOSE_ENV) goose validate
.PHONY: migration-validate

migration-fix: ### apply sequential ordering to migrations
	$(GOOSE_ENV) goose fix
.PHONY: migration-fix

# ==============================================================================
# Air Hot-Reload Targets
# ==============================================================================

air-init: ### initialize air configuration
	air init
.PHONY: air-init

air: ### run application with hot-reload
	air -c ./.air.toml
.PHONY: air

# ==============================================================================
# Tool Installation
# ==============================================================================

bin-deps: ### install development tools
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
.PHONY: bin-deps

keygen: ### generate fresh RSA key pair for JWT
	go run cmd/keygen/main.go
.PHONY: keygen

# ==============================================================================
# SQLC Code Generation
# ==============================================================================

sqlc-postgres: ### generate type-safe Go code from PostgreSQL queries
	cd internal/repo/persistent/postgres/sqlc && sqlc generate
.PHONY: sqlc-postgres

sqlc: sqlc-postgres ### generate all SQLC code
.PHONY: sqlc

pre-commit: swag mock format linter-golangci arch-check test ### run pre-commit checks
.PHONY: pre-commit

test-e2e: ### run e2e-test
	go clean -testcache && go test -v -count=1 -p 1 ./test/e2e/flows/...
.PHONY: test-e2e

# ==============================================================================
# Schemathesis API Testing
# ==============================================================================

API_URL ?= http://localhost:8080/api/v1
SCHEMA_URL ?= $(API_URL)/swagger/doc.json
SCHEMATHESIS_MAX_EXAMPLES ?= 50
SCHEMATHESIS_WORKERS ?= 4
VENV_DIR ?= .venv
PYTHON_VENV ?= $(VENV_DIR)/bin/python
SCHEMATHESIS_BIN ?= $(VENV_DIR)/bin/schemathesis

test-schemathesis-install: ### install schemathesis in venv
	@echo "📦 Creating virtual environment..."
	@python3 -m venv $(VENV_DIR)
	@echo "📦 Installing Schemathesis..."
	@$(PYTHON_VENV) -m pip install --no-cache-dir --upgrade pip
	@$(PYTHON_VENV) -m pip install --no-cache-dir schemathesis
	@echo "✅ Schemathesis installed successfully in $(VENV_DIR)!"
.PHONY: test-schemathesis-install

test-schemathesis-project: ### run schemathesis test on project schema (quick)
	@chmod +x test/schemathesis/run_tests_v2.sh
	@./test/schemathesis/run_tests_v2.sh
.PHONY: test-schemathesis-project

test-schemathesis: ### run schemathesis tests against local API
	@echo "🧪 Running Schemathesis tests..."
	@echo "📋 API URL: $(API_URL)"
	@echo "📋 Schema: docs/swagger/swagger.yaml"
	@if [ ! -f "$(SCHEMATHESIS_BIN)" ]; then \
		echo "❌ Schemathesis not found. Running install..."; \
		make test-schemathesis-install; \
	fi
	$(SCHEMATHESIS_BIN) run docs/swagger/swagger.yaml \
		--url="$(API_URL)" \
		--checks=all \
		--max-examples=$(SCHEMATHESIS_MAX_EXAMPLES) \
		--workers=$(SCHEMATHESIS_WORKERS) \
		--exclude-deprecated \
		| tee docs/report/schemathesis/report.txt
.PHONY: test-schemathesis

test-schemathesis-stateful: ### run schemathesis stateful tests (realistic workflows)
	@echo "🔄 Running Schemathesis stateful tests..."
	@if [ ! -f "$(SCHEMATHESIS_BIN)" ]; then \
		echo "❌ Schemathesis not found. Running install..."; \
		make test-schemathesis-install; \
	fi
	$(SCHEMATHESIS_BIN) run docs/swagger/swagger.yaml \
		--url="$(API_URL)" \
		--checks=all \
		--stateful=links \
		--max-examples=20 \
		--workers=2
.PHONY: test-schemathesis-stateful

test-schemathesis-quick: ### run quick schemathesis test (10 examples)
	@echo "⚡ Running quick Schemathesis test..."
	@if [ ! -f "$(SCHEMATHESIS_BIN)" ]; then \
		echo "❌ Schemathesis not found. Running install..."; \
		make test-schemathesis-install; \
	fi
	$(SCHEMATHESIS_BIN) run docs/swagger/swagger.yaml \
		--url="$(API_URL)" \
		--checks=all \
		--max-examples=10 \
		--exitfirst
.PHONY: test-schemathesis-quick

test-schemathesis-full: test-schemathesis test-schemathesis-stateful ### run all schemathesis tests
	@echo "✅ All Schemathesis tests completed!"
.PHONY: test-schemathesis-full

test-schemathesis-ci: ### run schemathesis tests for CI/CD
	@echo "🤖 Running Schemathesis tests for CI/CD..."
	@if [ ! -f "$(SCHEMATHESIS_BIN)" ]; then \
		echo "❌ Schemathesis not found. Running install..."; \
		make test-schemathesis-install; \
	fi
	$(SCHEMATHESIS_BIN) run "$(SCHEMA_URL)" \
		--url="$(API_URL)" \
		--checks=all \
		--max-examples=30 \
		--workers=4 \
		--exitfirst=false \
		--junit-xml=test-results/schemathesis.xml
.PHONY: test-schemathesis-ci

test-api-all: test test-e2e test-schemathesis ### run all API tests (unit, e2e, schemathesis)
	@echo "✅ All API tests completed!"
.PHONY: test-api-all

# ==============================================================================
# k6 Performance Testing
# ==============================================================================

K6_BASE_URL ?= http://localhost:8080

test-k6-smoke: ### run k6 smoke test (quick sanity)
	k6 run -e BASE_URL=$(K6_BASE_URL) test/performance/k6/scenarios/smoke.js
.PHONY: test-k6-smoke

test-k6-auth: ### run k6 auth flow load test
	k6 run -e BASE_URL=$(K6_BASE_URL) test/performance/k6/scenarios/auth-flow.js
.PHONY: test-k6-auth

test-k6-crud: ### run k6 CRUD load test
	k6 run -e BASE_URL=$(K6_BASE_URL) test/performance/k6/scenarios/crud-users.js
.PHONY: test-k6-crud

test-k6-files: ### run k6 file upload load test
	k6 run -e BASE_URL=$(K6_BASE_URL) test/performance/k6/scenarios/file-upload.js
.PHONY: test-k6-files

test-k6-mixed: ### run k6 mixed workload test
	k6 run -e BASE_URL=$(K6_BASE_URL) test/performance/k6/scenarios/mixed-workload.js
.PHONY: test-k6-mixed

test-k6-all: test-k6-smoke test-k6-auth test-k6-crud test-k6-files test-k6-mixed ### run all k6 tests
	@echo "✅ All k6 performance tests completed!"
.PHONY: test-k6-all

test-k6-ci: ### run k6 smoke test for CI (with JSON output)
	@mkdir -p test-results
	k6 run -e BASE_URL=$(K6_BASE_URL) --out json=test-results/k6-results.json test/performance/k6/scenarios/smoke.js
.PHONY: test-k6-ci

