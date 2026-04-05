# Kernel

Shared DDD building blocks and cross-cutting infrastructure used by all bounded contexts. Provides base types for entities, aggregates, domain events, and errors, plus a comprehensive infrastructure layer for databases, security, caching, observability, and more.

## Domain Building Blocks

### BaseEntity
- `BaseEntity` -- provides id (UUID), createdAt, updatedAt, deletedAt fields with auto-generated UUID on creation. Methods: `ID()`, `CreatedAt()`, `UpdatedAt()`, `DeletedAt()`, `IsDeleted()`, `Touch()`, `SoftDelete()`, `Restore()`.

### AggregateRoot
- `AggregateRoot` -- embeds `BaseEntity`, adds domain event recording. Methods: `AddEvent(event)`, `Events()`, `ClearEvents()`.

### DomainEvent
- `DomainEvent` interface -- `EventName() string`, `OccurredAt() time.Time`, `AggregateID() uuid.UUID`

### DomainError
- `DomainError` -- structured error with code and message. Supports `errors.Is()` matching by code.

### Generic Repository
- `Repository[T]` interface -- `Save(entity)`, `FindByID(id)`, `Update(entity)`, `Delete(id)`, `List(filter)`

### Value Objects
- `Pagination` -- Limit, Offset, Total, SortBy, SortOrder with getters/setters
- `SortOrder` -- ASC/DESC enum with `IsValid()` check
- `Lang` -- multi-language string (Uz, Ru, En)
- `File` -- file metadata (Name, Link)

### Constants (`domain/consts/`)
- `common.go` -- sort orders, date format, role names (user, admin, super_admin), auth bearer prefix
- `cache.go` -- cache key patterns
- `const.go` -- general constants
- `ctx.go` -- context key names
- `errors.go` -- shared error messages
- `header.go` -- HTTP header names
- `policy.go` -- authorization policy constants
- `repo.go` -- repository-level constants
- `response.go` -- standard response messages
- `router.go` -- route path constants
- `tables.go` -- database table names
- `use_case.go` -- use case constants

## Application Building Blocks

### CQRS Interfaces
- `CommandHandler[C]` -- `Handle(ctx, cmd) error`
- `QueryHandler[Q, R]` -- `Handle(ctx, query) (R, error)`

### Event Bus
- `EventBus` interface -- `Publish(ctx, events...)`, `Subscribe(eventName, handler)`
- `EventHandler` -- function type `func(ctx, DomainEvent) error`

## Infrastructure

### Database

| Package | Description |
|---------|-------------|
| `db/postgres` | PostgreSQL connection pool (pgxpool), migration runner (goose), LISTEN/NOTIFY listener, query helpers (Squirrel), OpenTelemetry tracer |
| `db/redis` | Redis client setup with functional options |
| `db/minio` | MinIO (S3-compatible) object storage client |
| `generic_repo` | Generic `BaseRepository[T]` with CRUD operations using Squirrel + pgxpool |
| `pgxutil` | `WithTx(pool, fn)` helper for running functions inside a database transaction |

### Security

| Package | Description |
|---------|-------------|
| `security/jwt` | RSA-based JWT TokenService: access/refresh token generation and validation, TokenPair DTO, device fingerprinting, secure token storage |
| `security/csrf` | CSRF protection with HMAC-SHA256 tokens, Double-Submit Cookie pattern, Redis-backed store, automatic token rotation |

### Caching (`cache/`)
In-memory cache implementations with TTL support:
- `MemoryCache` -- simple TTL cache with background cleanup
- `LRU` -- Least Recently Used eviction
- `LFU` -- Least Frequently Used eviction
- `MRU` -- Most Recently Used eviction
- `FIFO` -- First In First Out eviction
- `LIFO` -- Last In First Out eviction
- `SLRU` -- Segmented LRU (probation + protected segments)
- `TwoQueue` -- 2Q adaptive replacement cache
- `Random` -- random eviction

### Observability

| Package | Description |
|---------|-------------|
| `logger` | Structured logger interface (`Log`) with context-aware methods (Debugc, Infoc, Warnc, Errorc, Fatalc) and simple methods; Gin log writer adapter; color output support |
| `tracing` | OpenTelemetry tracer initialization with Jaeger exporter |

### Error Handling

| Package | Description |
|---------|-------------|
| `errors` | Comprehensive error system: typed error codes, PostgreSQL/Redis/MinIO error mapping, repository/service error factories, retry logic, error metrics, centralized error handler with HTTP response mapping |
| `errorx` | HTTP error response helpers, structured error logging service |

### Messaging and Tasks

| Package | Description |
|---------|-------------|
| `eventbus` | `InMemoryEventBus` -- in-memory implementation of `application.EventBus` for development and testing (synchronous, fail-fast) |
| `asynq` | Async task processing using Asynq (Redis-backed): client, worker, task definitions, handler registration |
| `firebase` | Firebase Cloud Messaging (FCM) push notification client with content builders |
| `telegram` | Telegram bot client for sending notifications (messages, alerts) |

### HTTP and Validation

| Package | Description |
|---------|-------------|
| `httpx` | HTTP utility layer: request binding, context helpers (user ID, role, session extraction), query/param parsing, file upload handling, security middleware, auth helpers, mock context for testing |
| `validation` | Field validators: email, phone, password (with fuzz testing), UUID, enum validation |
| `validator` | Struct validation wrapper |
| `server/http` | HTTP server with auto port-binding and figlet banner |

### Utilities

| Package | Description |
|---------|-------------|
| `contextx` | Type-safe context keys and helpers for request ID, session ID, user ID, role, IP, user agent, API version, trace ID |
| `useragent` | User-Agent string parser; detects device type (MOBILE, TABLET, DESKTOP, BOT) |
| `ptrutil` | Generic pointer utilities: `Ptr[T](v)`, `StrVal(*string)`, `IntVal(*int)`, `BoolVal(*bool)` |
| `featureflag` | Feature flag client using GoFeatureFlag with Redis retriever; includes Gin middleware for flag-gated endpoints |
| `container` | Test container helpers for PostgreSQL, Redis, and MinIO (testcontainers-go) |

## Usage
```go
import (
    "gct/internal/kernel/domain"
    "gct/internal/kernel/application"
    "gct/internal/kernel/infrastructure/logger"
    // ... other infrastructure packages as needed
)
```
