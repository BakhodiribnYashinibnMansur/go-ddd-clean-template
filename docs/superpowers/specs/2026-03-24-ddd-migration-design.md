# Backend DDD Migration — Design Spec

## Context

The current backend uses Clean Architecture with a layered structure (`controller → usecase → domain → repo`). While functional, the architecture doesn't enforce strong bounded context isolation, domain event-driven communication, or CQRS separation. This migration restructures the entire backend to Pure Domain-Driven Design to achieve:

- Strong bounded context isolation (20+ independent BCs)
- CQRS (Command Query Responsibility Segregation) in all BCs
- Event-driven inter-BC communication via Kafka + Protobuf
- Strict Aggregate Root pattern with Value Objects
- Anti-Corruption Layer where BC models diverge
- Layered testing strategy per DDD layer

**Migration strategy:** Big bang — full restructure in one pass, with a rollback plan (see Section 13).

---

## Decisions Summary

| Decision | Choice |
|---|---|
| Bounded Contexts | Each domain = separate BC (23 total) |
| Domain Events | Kafka + Protobuf serialization |
| CQRS | Full, in all BCs |
| Repository | Generic Repository[T] + domain-specific extensions |
| Folder Structure | Domain-centric (per BC folder) |
| Shared Infrastructure | Pure DDD — `pkg/` → `internal/shared/infrastructure/` |
| Error Handling | Existing system preserved, relocated to DDD layout |
| Value Objects | Full (Phone, Email, Password, etc.) |
| Aggregate Root | Strict (child entities accessed only through root) |
| ACL | Where BC models diverge; Shared Kernel / Published Language elsewhere |
| Domain Service | Domain Service (business rules) + Application Service (orchestration) |
| Test Strategy | Per-layer testing (unit → integration → E2E) |
| Migration | Big bang with git branch rollback plan |

---

## 1. Top-Level Folder Structure

```
Backend/
  cmd/                              # UNCHANGED
  config/                           # UNCHANGED
  consts/                           # → internal/shared/domain/consts/ (domain constants)
  migrations/                       # UNCHANGED
  docs/                             # UNCHANGED
  proto/                            # NEW — Protobuf event definitions
    events/
      user/v1/events.proto
      session/v1/events.proto
      authz/v1/events.proto
      audit/v1/events.proto
      ...
    buf.yaml                        # Buf toolchain config
    buf.gen.yaml                    # Protobuf code generation config
  internal/
    shared/                         # Shared Kernel + Infrastructure
      domain/                       # Shared domain model (base types)
      application/                  # Shared application interfaces
      infrastructure/               # pkg/ moves here entirely
    user/                           # Bounded Context
      domain/
      application/
      infrastructure/
      interfaces/
    session/                        # BC (User Aggregate child → own BC for queries only)
    authz/                          # BC
    audit/                          # BC
    dashboard/                      # BC
    featureflag/                    # BC
    integration/                    # BC
    webhook/                        # BC
    notification/                   # BC
    emailtemplate/                  # BC
    announcement/                   # BC
    translation/                    # BC
    sitesetting/                    # BC
    ratelimit/                      # BC
    iprule/                         # BC
    job/                            # BC
    dataexport/                     # BC
    file/                           # BC
    errorcode/                      # BC
    usersetting/                    # BC (passcode, user preferences)
    systemerror/                    # BC (error tracking, resolution workflow)
    metric/                         # BC (FunctionMetric — latency/panic tracking)
    app/                            # App bootstrap (wiring)
  test/                             # PRESERVED — existing test suites migrate in place
  go.mod
  Makefile
```

**What moves where:**

### pkg/ (22 packages) → `internal/shared/infrastructure/`
- `pkg/asynq/` → `internal/shared/infrastructure/asynq/`
- `pkg/broker/` → `internal/shared/infrastructure/broker/`
- `pkg/cache/` → `internal/shared/infrastructure/cache/`
- `pkg/container/` → `internal/shared/infrastructure/container/`
- `pkg/contextx/` → `internal/shared/infrastructure/contextx/`
- `pkg/csrf/` → `internal/shared/infrastructure/security/csrf/`
- `pkg/db/` → `internal/shared/infrastructure/postgres/`
- `pkg/errors/` → `internal/shared/infrastructure/errors/`
- `pkg/errorx/` → `internal/shared/infrastructure/errorx/`
- `pkg/featureflag/` → `internal/shared/infrastructure/featureflag/`
- `pkg/firebase/` → `internal/shared/infrastructure/firebase/`
- `pkg/httpx/` → `internal/shared/infrastructure/httpx/`
- `pkg/jwt/` → `internal/shared/infrastructure/security/jwt/`
- `pkg/logger/` → `internal/shared/infrastructure/logger/`
- `pkg/pgxutil/` → `internal/shared/infrastructure/pgxutil/`
- `pkg/ptrutil/` → `internal/shared/infrastructure/ptrutil/`
- `pkg/server/` → `internal/shared/infrastructure/server/`
- `pkg/telegram/` → `internal/shared/infrastructure/telegram/`
- `pkg/telemetry/` → `internal/shared/infrastructure/tracing/`
- `pkg/useragent/` → `internal/shared/infrastructure/useragent/`
- `pkg/validation/` → `internal/shared/infrastructure/validation/`
- `pkg/validator/` → `internal/shared/infrastructure/validator/`

### internal/ relocations
- `internal/domain/*.go` → split into each BC's `domain/`
- `internal/domain/common.go` → `internal/shared/domain/` (Lang, File, Pagination)
- `internal/domain/email.go` → `internal/shared/domain/email.go` (if empty, delete)
- `internal/domain/database.go` → `internal/shared/infrastructure/postgres/models.go` (DBSession, SlowQuery — infra monitoring)
- `internal/domain/mock/` → each BC's own `domain/mock/` or shared mock package
- `internal/usecase/*` → split into each BC's `application/command/` and `application/query/`
- `internal/usecase/usecase.go` → dissolves into `internal/app/app.go` (BC wiring)
- `internal/usecase/new.go` → dissolves into `internal/app/app.go` (UseCase constructor → BC factories)
- `internal/usecase/interfaces.go` → split into each BC's `domain/repository.go`
- `internal/usecase/log_action.go` → `internal/audit/application/command/log_action.go` (cross-cutting audit)
- `internal/usecase/minio/` → `internal/file/application/` (file operations BC)
- `internal/repo/repository.go` → dissolves into `internal/app/app.go`
- `internal/repo/persistent/persistent.go` → dissolves, each BC wires its own repo
- `internal/repo/persistent/postgres/*` → split into each BC's `infrastructure/postgres/`
- `internal/repo/persistent/redis/*` → `internal/shared/infrastructure/redis/` (shared infra)
- `internal/repo/persistent/minio/*` → `internal/shared/infrastructure/minio/`
- `internal/repo/persistent/mysql/*` → **DROP** (unused alternative DB, not needed for DDD migration)
- `internal/repo/integration/rest/` → `internal/shared/infrastructure/rest/`

### controller/ relocations
- `internal/controller/restapi/v1/*` → split into each BC's `interfaces/http/`
- `internal/controller/restapi/v1/minio/` → `internal/file/interfaces/http/` (file upload/download)
- `internal/controller/restapi/v1/admin/` → `internal/app/admin/` (admin utilities — linter, etc.)
- `internal/controller/restapi/v1/test/` → `internal/app/test/` (contract test endpoints — dev only)
- `internal/controller/restapi/v1/featureflagcrud/` → `internal/featureflag/interfaces/http/` (merges with featureflag BC)
- `internal/controller/restapi/controller.go` → dissolves into `internal/app/routes.go`
- `internal/controller/restapi/router.go` → `internal/app/routes.go`
- `internal/controller/restapi/setup_*.go` → `internal/app/` (setup_docs, setup_infra, setup_root, setup_admin_redirect)
- `internal/controller/restapi/middleware/` → `internal/shared/infrastructure/middleware/`
- `internal/controller/restapi/response/` → `internal/shared/infrastructure/httpx/response/`
- `internal/controller/restapi/cookie/` → `internal/shared/infrastructure/httpx/cookie/`

### Other relocations
- `internal/seeder/` → `internal/app/seeder/`
- `internal/cron/` → `internal/app/cron/`
- `internal/web/` → `internal/app/web/`
- `consts/` → `internal/shared/domain/consts/`
- `test/` — stays at top level, existing tests updated with new import paths

---

## 2. Bounded Context Internal Structure (User example)

```
internal/user/
  domain/
    entity.go              # User Aggregate Root (private fields, domain methods)
                           # Sessions are CHILD entities inside User aggregate
    session.go             # Session entity (belongs to User aggregate)
    value_object.go        # Phone, Email, Password (self-validating)
    repository.go          # UserRepository interface (extends Generic)
    event.go               # UserCreated, UserSignedIn, UserDeactivated
    error.go               # ErrUserNotFound, ErrPhoneExists, ErrInvalidPassword
    service.go             # SignInService, PasswordService (domain logic only)
  application/
    command/
      create_user.go       # CreateUserCommand + CreateUserHandler
      update_user.go
      delete_user.go
      sign_in.go           # Auth: sign in (creates session within User aggregate)
      sign_up.go           # Auth: sign up
      sign_out.go          # Auth: sign out (removes session from User aggregate)
      change_role.go
      approve_user.go
      bulk_action.go
    query/
      get_user.go          # GetUserQuery + GetUserHandler → UserView
      list_users.go
      search_users.go
    dto.go                 # UserView (read model), response DTOs
  infrastructure/
    postgres/
      write_repo.go        # UserRepository implementation (CQRS write)
      read_repo.go         # Read-optimized queries (CQRS read)
    kafka/
      producer.go          # Event publish (Protobuf serialization)
      consumer.go          # Event subscribe
    acl/
      authz_acl.go         # Translates Authz BC models → User BC models
  interfaces/
    http/
      handler.go           # Gin controller (delegates to command/query handlers)
      routes.go            # Route registration
  bc.go                    # BoundedContext factory + NewBoundedContext()
```

### Session clarification

Session is a **child entity** within the User Aggregate Root. It is NOT a separate aggregate root. The `session/` BC listed in Section 1 exists ONLY for read-side queries (CQRS Query side — "list all sessions", "get session by ID"). All write operations (create, revoke) go through the User aggregate:

```
Write: user.AddSession() → userRepo.Save(user)     # Through User aggregate
Read:  sessionQuery.ListByUserID(id) → []SessionView  # Separate read model
```

### Auth commands clarification

`internal/domain/auth.go` (SignIn, SignUp, SignOut, RefreshToken) maps to:
- `SignIn`, `SignUp`, `SignOut` → `user/application/command/` (User aggregate manages auth)
- `RefreshToken`, `RevokeSessions` → `user/application/command/` (Session is child of User)

---

## 3. Shared Kernel (internal/shared/)

**DDD clarification:** "Shared Kernel" here refers to the shared domain model subset (domain/ and application/). Infrastructure packages (postgres, redis, etc.) are separate — they are shared infrastructure, not part of the Shared Kernel in the DDD sense.

### domain/ (Shared Kernel proper)
- `base_entity.go` — BaseEntity: id, createdAt, updatedAt, deletedAt (soft delete)
- `aggregate_root.go` — AggregateRoot: embeds BaseEntity + events list, AddEvent(), ClearEvents()
- `event.go` — DomainEvent interface: EventName(), OccurredAt(), AggregateID()
- `repository.go` — Generic Repository[T] interface (see Section 3.1)
- `value_object.go` — Pagination, SortOrder, Lang (Uz/Ru/En), File (name/link)
- `error.go` — Base DomainError struct
- `consts/` — Domain constants (from top-level `consts/`)

### application/ (Shared application interfaces)
- `command.go` — Command interface, CommandHandler[C] interface
- `query.go` — Query interface, QueryResult interface, QueryHandler[Q, R] interface
- `event_bus.go` — EventBus interface: Publish(), Subscribe()

### infrastructure/ (Shared infrastructure — NOT Shared Kernel)

**From `pkg/` (22 packages):**
- `postgres/` — DB driver, connection pool (from `pkg/db/`)
- `redis/` — Redis client + bitmap, geospatial, hyperloglog, pubsub, stream, store (from `internal/repo/persistent/redis/`)
- `minio/` — MinIO S3 client (from `internal/repo/persistent/minio/`)
- `logger/` — Zap logger
- `errors/` — **FULLY PRESERVED** (AppError, 3-layer codes, factory, mapping, registry)
- `errorx/` — Error extensions
- `validator/` — Struct validation
- `validation/` — Additional validation utilities
- `security/` — JWT (from `pkg/jwt/`), CSRF (from `pkg/csrf/`)
- `httpx/` — HTTP utilities
- `tracing/` — OpenTelemetry (from `pkg/telemetry/`)
- `broker/` — Message broker abstractions
- `asynq/` — Asynq task queue client
- `cache/` — In-memory cache (LRU, LFU, FIFO, etc.) — infrastructure, NOT a BC
- `container/` — DI container utilities
- `contextx/` — Context extensions
- `useragent/` — User agent parser
- `ptrutil/` — Pointer utilities
- `pgxutil/` — PGX utilities
- `featureflag/` — Feature flag runtime utilities
- `firebase/` — Firebase client
- `telegram/` — Telegram bot client
- `server/` — HTTP server utilities

**From `internal/` (relocated to shared infra):**
- `rest/` — External REST API client (from `internal/repo/integration/rest/`)
- `middleware/` — HTTP middleware (from `internal/controller/restapi/middleware/`)
- `response/` — HTTP response formatters (from `internal/controller/restapi/response/`)
- `cookie/` — Cookie management (from `internal/controller/restapi/cookie/`)
- `kafka/` — Kafka client + NEW Protobuf EventBus implementation (from `pkg/broker/` + new code)
- `generic_repo.go` — NEW Generic Repository[T] base implementation (Squirrel)

### 3.1 Generic Repository[T] Interface

```go
type Repository[T any] interface {
    Save(ctx context.Context, entity *T) error
    FindByID(ctx context.Context, id uuid.UUID) (*T, error)
    Update(ctx context.Context, entity *T) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, filter Pagination) ([]*T, int64, error)
}

// Base implementation using Squirrel
type BaseRepository[T any] struct {
    db        *pgxpool.Pool
    tableName string
    columns   []string
    scanner   func(pgx.Row) (*T, error)  // entity-specific row scanner
}

func (r *BaseRepository[T]) Save(ctx context.Context, entity *T) error { ... }
func (r *BaseRepository[T]) FindByID(ctx context.Context, id uuid.UUID) (*T, error) { ... }
// ... etc
```

Go generics (1.22+) fully support this pattern. The `scanner` function handles entity-specific deserialization, avoiding reflection.

---

## 4. CQRS + Kafka + Protobuf Flow

### Write (Command) flow:
```
HTTP Request
  → interfaces/http/handler.go
    → application/command/handler.go
      → domain/service.go (business rules)
        → domain/entity.go (aggregate root adds events)
      → infrastructure/postgres/write_repo.go (save)
      → aggregate.ClearEvents() → EventBus.Publish()
        → Protobuf serialize → Kafka topic "{bc}.events.v1"
```

### Read (Query) flow:
```
HTTP Request
  → interfaces/http/handler.go
    → application/query/handler.go
      → infrastructure/postgres/read_repo.go (optimized query)
        → return DTO (UserView)
```

### Cross-BC event flow:
```
Kafka topic "user.events.v1"
  → audit/infrastructure/kafka/consumer.go
    → Protobuf deserialize
    → audit/infrastructure/acl/user_acl.go (translate)
      → audit/application/command/handler.go
        → audit/domain/ → audit/infrastructure/postgres/
```

### Kafka topic naming:
```
{bc_name}.events.v1       # user.events.v1, audit.events.v1
{bc_name}.commands.v1     # async commands (if needed)
```

### Kafka + Protobuf setup requirements:
- Add `kafka` service to `docker-compose.yml` (Confluent or Redpanda image)
- Add `buf` toolchain for Protobuf compilation (`buf.yaml`, `buf.gen.yaml`)
- Add `protoc-gen-go` for Go code generation
- Add Makefile target: `make proto` → generates Go structs from `.proto` files
- Event schema evolution: use Protobuf field numbering (never reuse field numbers)

---

## 5. Context Map & BC Relationships

### Context Map (DDD relationship types):

```
┌─────────────────────────────────────────────────────────┐
│                    CORE DOMAIN                           │
│  user (upstream) ──Published Language──→ audit           │
│  user (upstream) ──Published Language──→ notification    │
│  user (upstream) ──Published Language──→ dashboard       │
│  authz (upstream) ──Published Language──→ audit          │
│  authz (upstream) ──Shared Kernel──→ user               │
│                                                          │
│                    SUPPORTING                            │
│  session ──Conformist──→ user (read model of User child)│
│  webhook ──Customer-Supplier──→ integration              │
│  featureflag ──Published Language──→ all BCs (via event) │
│  sitesetting ──Published Language──→ all BCs (via event) │
│                                                          │
│                    GENERIC                               │
│  translation, emailtemplate, announcement                │
│  ratelimit, iprule, job, dataexport, file               │
│  errorcode, usersetting, systemerror, metric            │
│  (mostly independent, minimal cross-BC communication)    │
└─────────────────────────────────────────────────────────┘
```

### Relationship types used:
- **Published Language** — Upstream BC publishes Protobuf events. Downstream consumes. No ACL needed if both BCs agree on the event schema.
- **Shared Kernel** — `authz` ↔ `user` share Role/Permission concepts via `shared/domain/`
- **ACL** — Only where models **diverge**: e.g., `audit` translates `UserCreatedEvent` → `AuditActor` (different model)
- **Conformist** — `session` BC conforms to User aggregate's session model (read-only view)
- **Customer-Supplier** — `webhook` produces events, `integration` consumes and acts on them

### BC event communication map:
```
user ──event──→ audit (ACL), notification (PL), dashboard (PL)
authz ──event──→ audit (ACL), user (Shared Kernel)
session (via user) ──event──→ audit (ACL), dashboard (PL)
webhook ──event──→ integration (Customer-Supplier)
featureflag ──event──→ all subscribers (PL)
sitesetting ──event──→ all subscribers (PL)
job ──event──→ notification (PL)
dataexport ──event──→ notification (PL)
```

**Rule:** BCs communicate ONLY via Kafka events. No direct imports between BCs.

---

## 6. App Bootstrap (internal/app/)

- `app.go` — Main Run(): init infra → init shared → init BCs → register subscribers → register routes → start server
- `subscribers.go` — Kafka event subscription registration (which BC listens to which topic)
- `routes.go` — HTTP route registration (each BC registers its own routes)
- `middleware.go` — Middleware setup (preserved from current)
- `init_errors.go` — Error code initialization (preserved from current)
- `seeder/` — Database seeder (from `internal/seeder/`)
- `cron/` — Cron job scheduler (from `internal/cron/`)
- `web/` — Web system endpoints (from `internal/web/`)

**BC Factory pattern:**
```go
// internal/user/bc.go
func NewBoundedContext(pg *pgxpool.Pool, eventBus EventBus, logger Logger) *BoundedContext {
    repo := postgres.NewUserWriteRepo(pg)
    readRepo := postgres.NewUserReadRepo(pg)
    signInService := domain.NewSignInService()
    return &BoundedContext{
        CreateUser: command.NewCreateUserHandler(repo, eventBus, logger),
        SignIn:     command.NewSignInHandler(repo, signInService, eventBus, logger),
        GetUser:    query.NewGetUserHandler(readRepo, logger),
        ListUsers:  query.NewListUsersHandler(readRepo, logger),
    }
}

func (bc *BoundedContext) RegisterRoutes(api *gin.RouterGroup) {
    h := http.NewHandler(bc)
    users := api.Group("users")
    users.POST("", h.Create)
    users.GET("", h.List)
    users.GET("/:id", h.Get)
    ...
}
```

---

## 7. Test Strategy

| Layer | Test Type | Speed | Mocks | What's Tested |
|---|---|---|---|---|
| `domain/` | Unit | <1ms | None | Entity logic, VO validation, Domain Service rules |
| `application/` | Unit | <10ms | Repo, EventBus | Command/Query handler orchestration |
| `infrastructure/` | Integration | <5s | Real DB (testcontainers) | Postgres queries, Kafka produce/consume |
| `interfaces/` | E2E | <10s | Real server (httptest) | HTTP request → response cycle |

```bash
go test ./internal/user/domain/...                          # Domain only
go test ./internal/user/application/...                     # Application only
go test ./internal/user/infrastructure/... -tags=integration # Integration
go test ./internal/... -tags=integration                    # Everything
```

### Existing test migration:
- `test/` directory stays at top level
- `test/e2e/`, `test/integration/`, `test/contract/`, `test/performance/` — update import paths
- `test/functional/`, `test/schemathesis/` — update API endpoint paths if changed
- Existing per-usecase tests (e.g., `internal/usecase/announcement/create_test.go`) → move to `internal/announcement/application/command/create_test.go`
- Mock files (`internal/domain/mock/`) → split into each BC or use shared mock generation

---

## 8. Protobuf Event Definitions

```
proto/
  buf.yaml                            # Buf module config
  buf.gen.yaml                        # Code generation config (protoc-gen-go)
  events/
    user/v1/events.proto              # UserCreated, UserSignedIn, UserDeactivated
    session/v1/events.proto           # SessionCreated, SessionRevoked (via User aggregate)
    authz/v1/events.proto             # RoleChanged, PolicyUpdated, PermissionGranted
    audit/v1/events.proto             # AuditLogCreated
    featureflag/v1/events.proto       # FlagToggled
    sitesetting/v1/events.proto       # SettingUpdated
    notification/v1/events.proto      # NotificationSent
    webhook/v1/events.proto           # WebhookTriggered
    job/v1/events.proto               # JobScheduled, JobCompleted
    dataexport/v1/events.proto        # ExportRequested, ExportCompleted
    metric/v1/events.proto            # FunctionMetricRecorded
```

Generated Go code goes to: `internal/shared/infrastructure/kafka/gen/`

---

## 9. Error Handling (Preserved)

The existing 3-layer error system is **fully preserved**:
- `AppError` struct with severity, category, retryable, suggestion
- Repo (2xxx) → Service (3xxx) → Handler (4xxx/5xxx) error codes
- `MapRepoToServiceError()`, `MapServiceToHandlerError()` mapping
- Fluent builder pattern (`.WithField().WithInput()`)
- Dynamic error code loading from DB
- System error middleware

**Only change:** File location moves from `pkg/errors/` → `internal/shared/infrastructure/errors/`

Each BC adds its own domain errors in `{bc}/domain/error.go` using the existing `apperrors` package.

---

## 10. Complete BC List (23)

| BC | Aggregate Root | Key Events | Type |
|---|---|---|---|
| user | User (+ Sessions, Attributes) | UserCreated, UserSignedIn, UserDeactivated | Core |
| session | — (read-only view of User's sessions) | — (events via User) | Supporting |
| authz | Role (+ Permission, Scope, Policy) | RoleChanged, PolicyUpdated | Core |
| audit | AuditLog | AuditLogCreated | Supporting |
| dashboard | — (read-only aggregation) | — | Supporting |
| featureflag | FeatureFlag | FlagToggled | Supporting |
| integration | Integration | IntegrationConnected | Supporting |
| webhook | Webhook | WebhookTriggered | Supporting |
| notification | Notification | NotificationSent | Generic |
| emailtemplate | EmailTemplate | TemplateUpdated | Generic |
| announcement | Announcement | AnnouncementPublished | Generic |
| translation | Translation | TranslationUpdated | Generic |
| sitesetting | SiteSetting | SettingUpdated | Generic |
| ratelimit | RateLimit | RateLimitChanged | Generic |
| iprule | IPRule | IPRuleCreated | Generic |
| job | Job | JobScheduled, JobCompleted | Generic |
| dataexport | DataExport | ExportRequested, ExportCompleted | Generic |
| file | FileMetadata | FileUploaded | Generic |
| errorcode | ErrorCode | ErrorCodeUpdated | Generic |
| usersetting | UserSetting | UserSettingChanged | Generic |
| systemerror | SystemError | SystemErrorRecorded, ErrorResolved | Supporting |
| metric | FunctionMetric | FunctionMetricRecorded | Supporting |

**Note:** `cache` and `database` are NOT BCs — they are infrastructure concerns:
- Cache (LRU/LFU/etc.) stays in `internal/shared/infrastructure/cache/`
- Database health/info stays in `internal/app/` as an infrastructure endpoint

---

## 11. Ubiquitous Language (per BC)

### User BC
- **User** — registered person with phone, email, credentials
- **Session** — active device connection (child of User)
- **Phone/Email/Password** — validated identity credentials (Value Objects)
- **Sign In/Up/Out** — authentication lifecycle
- **Approve** — admin enables pending user
- **Deactivate** — soft-disable user access

### Authz BC
- **Role** — named permission group (Aggregate Root)
- **Permission** — hierarchical access right (child of Role)
- **Scope** — API endpoint (path + method)
- **Policy** — Casbin rule (ALLOW/DENY effect)

### Audit BC
- **AuditLog** — recorded user action (Aggregate Root)
- **AuditActor** — who performed action (translated from User via ACL)
- **EndpointHistory** — API hit record
- **SystemError** → separate BC (systemerror)

### Other BCs
Each BC defines its own language in `{bc}/domain/` through entity names, method names, and error messages.

---

## 12. Kafka + Protobuf Infrastructure Setup

### docker-compose.yml additions:
```yaml
kafka:
  image: confluentinc/cp-kafka:7.6.0
  ports:
    - "9092:9092"
  environment:
    KAFKA_KRAFT_MODE: "true"
    # ... KRaft config (no Zookeeper)

schema-registry:  # optional, for Protobuf schema evolution
  image: confluentinc/cp-schema-registry:7.6.0
```

### Makefile additions:
```makefile
proto:  ## Generate Go code from .proto files
	buf generate

proto-lint:  ## Lint proto files
	buf lint
```

### go.mod additions:
```
google.golang.org/protobuf
github.com/segmentio/kafka-go  # already exists
```

### CI/CD changes:
- Add `buf` installation step
- Add `make proto` to build pipeline
- Add Kafka to test infrastructure (testcontainers)

---

## 13. Rollback Plan

Since this is a big bang migration:

1. **Before starting:** Create a new git branch `feat/ddd-migration`
2. **Main branch untouched** — all work on the feature branch
3. **If migration fails:** Simply switch back to main branch
4. **Verification gates before merge:**
   - All existing tests pass with new import paths
   - `go build ./...` succeeds
   - `go vet ./...` passes
   - Application starts and responds to health check
   - Manual smoke test: sign in, create user, list users
5. **Post-merge monitoring:** Keep old branch for 1 week as rollback reference
