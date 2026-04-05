# GCA — Go Clean Architecture Backend

Production-ready Go backend built with **Domain-Driven Design**, **CQRS**, and **Hexagonal Architecture**.

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.26 |
| Web Framework | Gin |
| Database | PostgreSQL 18 |
| Query Builder | Squirrel |
| Cache | Redis 7.4 (go-redis) |
| Task Queue | Asynq (Redis-backed) |
| File Storage | MinIO (S3-compatible) |
| Logging | Zap (structured) |
| Metrics | OpenTelemetry + Prometheus |
| Tracing | OpenTelemetry + Jaeger |
| Auth | JWT (RSA) + CSRF (Double-Submit Cookie) |
| API Docs | Swagger (swaggo/swag) |
| Migrations | Goose v3 |
| Testing | testify, testcontainers, schemathesis |

## Quick Start

```bash
# 1. Install dev tools
make bin-deps

# 2. Copy env and configure
cp .env.example .env

# 3. Generate JWT RSA keys
make keygen

# 4. Start PostgreSQL & Redis
make compose-up

# 5. Run migrations
make migration-up

# 6. Run the application
make run
```

The API will be available at `http://localhost:8080`. Swagger docs at `/api/v1/swagger/index.html`.

## Project Structure

```
Backend/
├── cmd/
│   ├── app/               # Main application entry point
│   ├── migration/          # Database migration runner
│   ├── seeder/             # Database seeding
│   └── keygen/             # JWT RSA key generation
│
├── config/                 # Configuration (env + YAML)
├── migrations/postgres/    # Goose SQL migrations
│
├── internal/
│   ├── app/                # Bootstrap, wiring, routes
│   ├── kernel/             # Shared kernel (see below)
│   ├── contract/           # Cross-BC contracts (events, ports)
│   └── context/            # Bounded Contexts (hybrid area → tier → BC layout)
│       ├── iam/            #   Identity & Access Management
│       ├── ops/            #   Operational concerns
│       ├── content/        #   Content & messaging
│       └── admin/          #   Admin / platform-level concerns
│
├── test/
│   ├── e2e/                # End-to-end flow tests
│   ├── integration/        # Integration tests (testcontainers)
│   ├── performance/        # Load & stress tests
│   └── schemathesis/       # API schema fuzzing
│
├── docs/swagger/           # Generated OpenAPI specs
├── nginx/                  # Reverse proxy config
├── Dockerfile              # Multi-stage (distroless)
└── docker-compose.yml      # Full infrastructure stack
```

## Architecture

### Design Principles

This project implements three complementary architectural patterns:

**Domain-Driven Design (DDD)** — The codebase is organized around business domains (Bounded Contexts), not technical layers. Each BC owns its entities, rules, and data access independently. BCs communicate through domain events, not direct imports — this keeps them decoupled and independently deployable.

**Hexagonal Architecture (Ports & Adapters)** — Each BC defines its contracts as interfaces (ports) in the domain layer. Infrastructure (PostgreSQL, Redis, HTTP) implements these interfaces as adapters. This means you can swap PostgreSQL for MongoDB, or REST for gRPC, without touching business logic.

**CQRS (Command Query Responsibility Segregation)** — Read and write operations are separated into distinct handlers with separate repository interfaces. Commands mutate state and may emit domain events. Queries are optimized for read performance. This separation allows independent scaling and optimization of read vs write paths.

### Bounded Context Structure

BCs live under `internal/context/<area>/<tier>/<bc>/` where `area ∈ {iam, ops, content, admin}` and `tier ∈ {core, supporting, generic}`. The tier reflects the BC's strategic classification (DDD Blue Book, Part IV) and is visible in the import path. Reclassification = `git mv` between tier sub-folders within the same area. See [docs/architecture/context-map.md](docs/architecture/context-map.md) for the single source of truth on classification.

```
internal/context/<area>/<tier>/<bc>/
│
├── domain/                     # DOMAIN LAYER (innermost — no external dependencies)
│   ├── entity.go              # Aggregate Root with business rules and invariants
│   ├── repository.go          # Port interfaces (WriteRepository, ReadRepository)
│   ├── event.go               # Domain events (e.g., UserCreated, OrderPlaced)
│   ├── error.go               # Domain-specific errors
│   └── value_object.go        # Immutable value types (Email, Money, Address)
│
├── application/                # APPLICATION LAYER (orchestrates domain objects)
│   ├── command/               # Write side — each handler = one use case
│   │   ├── create_handler.go  # Validates input → calls domain → persists → emits events
│   │   ├── update_handler.go
│   │   └── delete_handler.go
│   └── query/                 # Read side — optimized for data retrieval
│       ├── get_handler.go     # Single entity fetch with projections
│       └── list_handler.go    # Paginated list with filters and sorting
│
├── infrastructure/             # INFRASTRUCTURE LAYER (adapters — implements ports)
│   ├── postgres/
│   │   ├── write_repo.go      # Implements WriteRepository using Squirrel query builder
│   │   └── read_repo.go       # Implements ReadRepository with optimized queries
│   └── acl/                   # Anti-Corruption Layer for cross-BC communication
│       └── user_acl.go        # Translates external BC models to local domain models
│
├── interfaces/                 # INTERFACE LAYER (how the outside world talks to this BC)
│   └── http/
│       ├── handler.go         # Gin HTTP handlers — thin, delegates to application layer
│       └── middleware/        # BC-specific middleware (e.g., auth, signature verification)
│
└── bc.go                       # Bounded Context wiring — constructs all layers via DI
```

### How the Layers Interact

```
HTTP Request
    │
    ▼
┌─────────────────────────────────────────────────────────────────┐
│  INTERFACES (handler.go)                                        │
│  Parses HTTP request → builds Command/Query DTO → calls app    │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  APPLICATION (command/ or query/)                                │
│  Validates business rules → orchestrates domain objects          │
│  Commands: entity.Create() → repo.Save() → eventBus.Publish()  │
│  Queries: repo.FindByID() → return DTO                          │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  DOMAIN (entity.go, repository.go)                              │
│  Pure business logic — no frameworks, no I/O                    │
│  Defines WHAT the system does, not HOW                          │
└──────────────────────────┬──────────────────────────────────────┘
                           │ (interface)
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  INFRASTRUCTURE (postgres/write_repo.go)                        │
│  Implements repository interfaces with actual database calls    │
│  Translates domain objects ↔ database rows                      │
└─────────────────────────────────────────────────────────────────┘
```

### Dependency Rule

Dependencies always point **inward**:

```
Interfaces → Application → Domain ← Infrastructure
```

- **Domain** depends on nothing (pure Go, no imports from other layers)
- **Application** depends only on Domain interfaces
- **Infrastructure** implements Domain interfaces (depends inward)
- **Interfaces** calls Application handlers (depends inward)

Infrastructure depends on Domain, not the other way around. This is the **Dependency Inversion Principle** — the domain defines what it needs (ports), and infrastructure provides it (adapters).

### Domain Events & Cross-BC Communication

```
┌──────────┐   CommandHandler    ┌──────────┐   EventBus.Publish()   ┌──────────────┐
│  User BC │ ──── creates ────→ │ UserCreated│ ──── publishes ─────→ │ Notification │
│          │      user           │   Event   │                        │  BC handles  │
└──────────┘                     └──────────┘                        └──────────────┘
```

BCs never import each other directly. When `User BC` creates a user, it publishes a `UserCreated` domain event. `Notification BC` subscribes to this event and sends a welcome email. This keeps BCs decoupled — you can remove `Notification BC` without breaking `User BC`.

### CQRS in Practice

```
                    ┌─────────── Command Path (Write) ───────────┐
                    │                                              │
HTTP POST/PUT/DEL → CommandHandler → Domain Entity → WriteRepo → DB
                    │                     │
                    │               Emits DomainEvent
                    │                     │
                    │               EventBus → Subscribers
                    │
                    ├─────────── Query Path (Read) ──────────────┐
                    │                                              │
HTTP GET ─────────→ QueryHandler → ReadRepo → DB (optimized SQL)
```

Write and read repositories have separate interfaces. Write repos enforce domain invariants. Read repos use optimized SQL (joins, projections) without loading full aggregates.

### Bounded Contexts (Area × Tier)

Strategic tier classification (Core / Supporting / Generic) with full justification: [docs/architecture/context-map.md](docs/architecture/context-map.md). The **Core** tier is intentionally empty in this template — product-specific Core BCs live under `<area>/core/` (or a new area) when the template is forked. Each area reserves a `core/.gitkeep` slot.

**iam — Identity & Access Management**

| BC | Tier | Location | Purpose |
|----|------|----------|---------|
| `user` | generic | `iam/generic/user` | User management, authentication (SignIn, SignUp, CRUD) |
| `session` | generic | `iam/generic/session` | Session management with device tracking |
| `authz` | generic | `iam/generic/authz` | RBAC — Role, Permission, Policy, Scope |
| `usersetting` | generic | `iam/generic/usersetting` | Per-user preferences |
| `audit` | supporting | `iam/supporting/audit` | Audit logging (GDPR/SOC2 compliance) |

**ops — Operational Concerns**

| BC | Tier | Location | Purpose |
|----|------|----------|---------|
| `metric` | generic | `ops/generic/metric` | Performance metrics collection |
| `ratelimit` | generic | `ops/generic/ratelimit` | Rate limiting rules |
| `systemerror` | generic | `ops/generic/systemerror` | Automatic 5xx error tracking |
| `iprule` | supporting | `ops/supporting/iprule` | IP-based access control (whitelist/blacklist) |

**content — Content & Messaging**

| BC | Tier | Location | Purpose |
|----|------|----------|---------|
| `notification` | generic | `content/generic/notification` | User notifications (in-app, push) |
| `file` | generic | `content/generic/file` | File metadata management (MinIO backend) |
| `translation` | generic | `content/generic/translation` | Multi-language support (i18n) |
| `announcement` | supporting | `content/supporting/announcement` | System announcements (multilingual) |

**admin — Admin / Platform**

| BC | Tier | Location | Purpose |
|----|------|----------|---------|
| `featureflag` | generic | `admin/generic/featureflag` | Feature toggles with Redis/file-based evaluation |
| `statistics` | supporting | `admin/supporting/statistics` | Business KPI aggregations |
| `integration` | supporting | `admin/supporting/integration` | External system integrations with API keys |
| `sitesetting` | supporting | `admin/supporting/sitesetting` | Site-wide configuration |
| `dataexport` | supporting | `admin/supporting/dataexport` | Async data export jobs (GDPR portability) |
| `errorcode` | supporting | `admin/supporting/errorcode` | Error code catalog (public API contract) |

### Shared Kernel (`internal/kernel/`)

Cross-cutting infrastructure used by all bounded contexts:

```
kernel/
├── domain/consts/          # Shared constants (claim names, header keys)
├── application/            # EventBus interface, CQRS base types
└── infrastructure/
    ├── db/postgres/        # PostgreSQL connection pool (pgxpool) + Squirrel query builder
    ├── db/redis/           # Redis client
    ├── db/minio/           # MinIO (S3-compatible) object storage client
    ├── security/jwt/       # JWT generation & validation (RSA) + device fingerprinting
    ├── security/csrf/      # CSRF double-submit cookie (HMAC-SHA256, Redis-backed)
    ├── logger/             # Zap structured logging
    ├── tracing/            # OpenTelemetry distributed tracing (Jaeger exporter)
    ├── metrics/            # OpenTelemetry + Prometheus metrics provider
    ├── asynq/              # Background job queue (Redis-backed)
    ├── cache/              # In-memory caches (LRU, LFU, SLRU, 2Q, FIFO, LIFO, MRU, Random)
    ├── eventbus/           # In-memory domain event bus
    ├── errors/             # Typed error codes, DB error mapping, retry logic
    ├── errorx/             # HTTP error response helpers
    ├── httpx/              # HTTP request/response utilities, context helpers
    ├── middleware/          # CORS, security headers, rate limiting
    ├── sse/                # Server-Sent Events
    ├── pubsub/             # Pub/sub messaging
    ├── metadata/           # Entity metadata (EAV key-value store)
    ├── validator/          # Struct validation wrapper
    ├── validation/         # Field validators (email, phone, password, UUID)
    ├── contextx/           # Type-safe context keys (request ID, user ID, role, etc.)
    ├── pgxutil/            # Transaction helpers (WithTx)
    ├── ptrutil/            # Generic pointer utilities
    ├── useragent/          # User-Agent parser (device type detection)
    ├── server/http/        # HTTP server with auto port-binding
    ├── container/          # Test container helpers (testcontainers-go)
    ├── firebase/           # Firebase Cloud Messaging (push notifications)
    └── telegram/           # Telegram bot integration
```

## Security

- **Authentication**: RSA-signed JWT (access + refresh tokens)
- **Authorization**: RBAC with per-endpoint permission checks
- **CSRF**: Double-submit cookie pattern
- **Rate Limiting**: Redis-backed sliding window (configurable)
- **Security Headers**: HSTS, CSP, X-Frame-Options, X-Content-Type-Options
- **CORS**: Configurable origin validation with wildcard support
- **Input Validation**: Strict JSON parsing with 2MB body limit

## Observability

- **Logging**: Structured JSON logging via Zap with log levels
- **Metrics**: OpenTelemetry + Prometheus exporter (`/metrics` endpoint)
- **Tracing**: OpenTelemetry spans + Jaeger exporter
- **Health Checks**: `/health/live`, `/health/ready`, `/healthz`, `/ping`
- **System Errors**: Automatic 5xx error persistence to database

## Testing

```bash
make test                      # Unit tests with race detection & coverage
make test-e2e                  # End-to-end flow tests
make test-fuzz                 # Fuzz testing
make test-prop                 # Property-based tests
make test-schemathesis         # API schema fuzzing (OpenAPI)
make test-schemathesis-stateful # Stateful workflow tests
make test-api-all              # Run all test suites
```

## Development

```bash
make air                       # Hot-reload with Air
make swag                      # Regenerate Swagger docs
make mock                      # Regenerate mocks
make format                    # Format code (gofumpt + gci)
make linter-golangci           # Lint with golangci-lint
make deps-audit                # Check for known vulnerabilities
make pre-commit                # Run all checks (swag, mock, format, lint, test)
```

## Database Migrations

```bash
make migration-create          # Create new migration (interactive)
make migration-up              # Run all pending migrations
make migration-down            # Rollback 1 migration
make migration-status          # Check migration status
make migration-reset           # Rollback all migrations
make migration-validate        # Validate without running
```

## Docker

```bash
make compose-up                # Start PostgreSQL & Redis
make compose-up-all            # Start full stack (backend + nginx + infra)
make compose-down              # Tear down all containers
```

`docker-compose.yml` includes: PostgreSQL, Redis, MinIO, RabbitMQ, NATS, Elasticsearch, ClickHouse, Cassandra, MongoDB, Jaeger.

## Configuration

All configuration is managed via environment variables (`.env`) with YAML defaults (`config/config.yaml`). See [.env.example](.env.example) for all available options.

Key configuration groups:
- **App**: `APP_NAME`, `APP_VERSION`
- **HTTP**: `HTTP_PORT` (default: 8080)
- **Database**: `PG_HOST`, `PG_PORT`, `PG_NAME`, `PG_USER`, `PG_PASSWORD`, `PG_POOL_MAX`
- **Redis**: `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`
- **JWT**: `JWT_PRIVATE_KEY`, `JWT_PUBLIC_KEY`, `JWT_ACCESS_TTL`, `JWT_REFRESH_TTL`
- **Observability**: `METRICS_ENABLED`, `TRACING_ENABLED`, `LOG_LEVEL`
- **Feature Flags**: `FEATURE_FLAG_ENABLED`, `FEATURE_FLAG_USE_REDIS`
- **Rate Limiting**: `LIMITER_ENABLED`, `LIMITER_LIMIT`, `LIMITER_PERIOD`

## Bootstrap Flow

```
cmd/app/main.go → config.NewConfig() → app.Run(cfg)
  1. Logger (Zap)
  2. Telemetry (OpenTelemetry + Jaeger)
  3. PostgreSQL connection pool (pgx)
  4. Redis connection
  5. Asynq task queue client
  6. DDD: NewDDDBoundedContexts() — instantiate all BCs
  7. EventBus (InMemory)
  8. Cache invalidation service
  9. Asynq Worker (background jobs)
 10. HTTP Router (Gin) + Middleware + Routes
 11. Graceful Shutdown (SIGINT/SIGTERM)
```

## CI/CD

GitHub Actions workflow (`.github/workflows/api-tests.yml`):
- Runs on push to `main`/`develop`/`master` and pull requests
- PostgreSQL + Redis service containers with health checks
- Schemathesis API schema fuzzing

## License

MIT
