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
	$(BASE_STACK) up --build -d db rabbitmq nats && docker compose logs -f
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

protogen: ### generate all proto files
	@bash script/protogen.sh
.PHONY: protogen

lint-proto: ### lint proto files using buf
	cd docs/protobuf/proto && buf lint
.PHONY: lint-proto

check-breaking: ### check breaking changes
	cd docs/protobuf/proto && buf breaking --against "../../.git#branch=master,subdir=docs/protobuf/proto"
.PHONY: check-breaking

doc-proto: ### generate proto documentation (all proto files in single HTML)
	mkdir -p docs/protobuf/doc
	protoc --doc_out=docs/protobuf/doc --doc_opt=html,index.html \
		--proto_path=docs/protobuf/proto \
		docs/protobuf/proto/v1/common/pagination.proto \
		docs/protobuf/proto/v1/user/user.proto \
		docs/protobuf/proto/v1/user/session.proto
.PHONY: doc-proto

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

run: ### run application with all code generation
	swag init --parseDependency --parseInternal -g cmd/app/main.go -o docs/swagger > /dev/null 2>&1 && \
	CGO_ENABLED=0 go run ./cmd/app
.PHONY: run

docker-rm-volume: ### remove docker volume
	docker volume rm go-clean-template_pg-data
.PHONY: docker-rm-volume

linter-golangci: ### check by golangci linter
	golangci-lint run
.PHONY: linter-golangci

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
# Mongo Migration Targets
# ==============================================================================

export MIGRATE_MONGO_DIR = $(CURRENT_DIR)/migrations/mongo
export MONGO_DSN ?= "mongodb://localhost:27017/test"

mongo-migration-create: ### create new mongo migration
	@read -p "Enter migration name: " MIGRATION_NAME; \
	migrate create -ext json -dir $(MIGRATE_MONGO_DIR) -seq $$MIGRATION_NAME
.PHONY: mongo-migration-create

mongo-migration-up: ### run mongo migrations
	migrate -path $(MIGRATE_MONGO_DIR) -database "$(MONGO_DSN)" up
.PHONY: mongo-migration-up

mongo-migration-down: ### roll back mongo migration
	migrate -path $(MIGRATE_MONGO_DIR) -database "$(MONGO_DSN)" down 1
.PHONY: mongo-migration-down

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
	go install -tags 'mongodb' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install github.com/air-verse/air@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
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

sqlc-mysql: ### generate type-safe Go code from MySQL queries
	cd internal/repo/persistent/mysql/sqlc && sqlc generate
.PHONY: sqlc-mysql

sqlc-sqlite: ### generate type-safe Go code from SQLite queries
	cd internal/repo/persistent/sqlite/sqlc && sqlc generate
.PHONY: sqlc-sqlite

sqlc: sqlc-postgres sqlc-mysql sqlc-sqlite ### generate all SQLC code
.PHONY: sqlc

pre-commit: swag protogen mock format linter-golangci test ### run pre-commit checks
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

