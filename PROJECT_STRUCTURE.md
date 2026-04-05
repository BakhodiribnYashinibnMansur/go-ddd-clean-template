# Golang Template — DDD Architecture

## Umumiy Ko'rinish

```
Backend/
├── cmd/                           # Dastur kirish nuqtalari
│   ├── app/main.go               # Asosiy dastur
│   ├── seeder/                   # DB seed qilish
│   ├── keygen/                   # JWT RSA kalit generatsiya
│   └── migration/                # Migratsiya ishga tushirish
│
├── config/                        # Konfiguratsiya
│   ├── config.go                 # Config struct & NewConfig()
│   ├── config.yaml               # YAML konfiguratsiya
│   ├── database.go               # DB sozlamalari
│   ├── jwt.go                    # JWT sozlamalari
│   └── redis.go                  # Redis sozlamalari
│
├── migrations/postgres/           # Goose SQL migratsiyalar
│
├── internal/                      # Asosiy business logika
│   ├── app/                      # Bootstrap & Wiring
│   ├── kernel/                   # Shared Kernel (infrastructure, domain, application primitives)
│   ├── contract/                 # BC contract slots (events, ports) — cross-BC contracts
│   └── context/                  # Bounded Contexts (hybrid area → tier → BC)
│       ├── iam/                  #   Identity & Access Management
│       ├── ops/                  #   Operational concerns
│       ├── content/              #   Content & messaging
│       └── admin/                #   Admin / platform-level concerns
│
├── test/                          # Testlar
│   ├── e2e/                      # End-to-end flow testlar
│   ├── integration/              # Integration testlar (testcontainers)
│   ├── performance/              # Load & stress testlar
│   └── schemathesis/             # API schema fuzzing
│
├── docs/swagger/                  # Generated OpenAPI specs
├── nginx/                         # Nginx config
├── Dockerfile                     # Multi-stage (distroless)
├── docker-compose.yml             # To'liq infra stack
└── Makefile
```

---

## DDD Bounded Context Strukturasi

Har bir Bounded Context `internal/context/<area>/<tier>/<bc>/` yo'lida joylashadi, bu yerda:
- **area** ∈ {`iam`, `ops`, `content`, `admin`} — domen hududi (cohesion saqlanadi)
- **tier** ∈ {`core`, `supporting`, `generic`} — strategik tier (DDD Blue Book, Part IV)

Tier reklassifikatsiya — `git mv` bilan shu area ichida tier sub-folderlar o'rtasida. Classification'ning yagona manbai: [docs/architecture/context-map.md](docs/architecture/context-map.md).

Har bir BC quyidagi hexagonal/onion arxitektura layerlaridan iborat:

```
internal/context/<area>/<tier>/<bc>/
├── domain/                        # DOMAIN LAYER
│   ├── entity.go                 # Aggregate Root / Entity
│   ├── repository.go             # Repository interfeysi (port)
│   ├── event.go                  # Domain eventlar
│   ├── error.go                  # Domain xatoliklar
│   └── value_object.go           # Value Objectlar
│
├── application/                   # APPLICATION LAYER (CQRS)
│   ├── command/                  # Write operatsiyalar
│   │   ├── create_handler.go
│   │   ├── update_handler.go
│   │   └── delete_handler.go
│   └── query/                    # Read operatsiyalar
│       ├── get_handler.go
│       └── list_handler.go
│
├── infrastructure/                # INFRASTRUCTURE LAYER (adapter)
│   └── postgres/
│       ├── write_repo.go         # WriteRepository implementatsiyasi
│       └── read_repo.go          # ReadRepository implementatsiyasi
│
├── interfaces/                    # INTERFACES LAYER
│   └── http/
│       └── handler.go            # REST API endpointlar
│
└── bc.go                          # Bounded Context wiring (DI)
```

---

## Bounded Context ro'yxati (Area × Tier)

> Strategik tier tasnifi (Core / Supporting / Generic) — to'liq asoslar bilan: [docs/architecture/context-map.md](docs/architecture/context-map.md).
> Template holatida **Core** tier bo'sh (har area'da `core/.gitkeep` slot zahirada) — mahsulot core BC'lari fork qilinganda tegishli `<area>/core/` ichiga qo'shiladi.

### 🔐 iam — Identity & Access Management
| BC | Tier | Location | Vazifasi |
|----|------|----------|----------|
| **user** | generic | `iam/generic/user` | Foydalanuvchi boshqaruvi, auth (SignIn, SignUp, CRUD) |
| **session** | generic | `iam/generic/session` | Sessiya boshqaruvi, device tracking |
| **authz** | generic | `iam/generic/authz` | RBAC (Role, Permission, Policy, Scope) |
| **usersetting** | generic | `iam/generic/usersetting` | Foydalanuvchi sozlamalari |
| **audit** | supporting | `iam/supporting/audit` | Audit logging (GDPR/SOC2 compliance) |

### ⚙️ ops — Operational Concerns
| BC | Tier | Location | Vazifasi |
|----|------|----------|----------|
| **metric** | generic | `ops/generic/metric` | Performance metrikalari |
| **ratelimit** | generic | `ops/generic/ratelimit` | Rate limiting qoidalari |
| **systemerror** | generic | `ops/generic/systemerror` | Tizim xatolarini kuzatish (5xx) |
| **iprule** | supporting | `ops/supporting/iprule` | IP asosida kirish nazorati |

### 📣 content — Content & Messaging
| BC | Tier | Location | Vazifasi |
|----|------|----------|----------|
| **notification** | generic | `content/generic/notification` | Foydalanuvchi bildirishnomalari |
| **file** | generic | `content/generic/file` | Fayl metadata boshqaruvi (MinIO) |
| **translation** | generic | `content/generic/translation` | Ko'p tilli qo'llab-quvvatlash (i18n) |
| **announcement** | supporting | `content/supporting/announcement` | Tizim e'lonlari (ko'p tilli) |

### 🛠 admin — Admin / Platform
| BC | Tier | Location | Vazifasi |
|----|------|----------|----------|
| **featureflag** | generic | `admin/generic/featureflag` | Feature toggle |
| **statistics** | supporting | `admin/supporting/statistics` | Business KPI agregatsiyalari |
| **integration** | supporting | `admin/supporting/integration` | Tashqi tizim integratsiyalari, API kalitlar |
| **sitesetting** | supporting | `admin/supporting/sitesetting` | Sayt konfiguratsiyasi |
| **dataexport** | supporting | `admin/supporting/dataexport` | Ma'lumot eksport ishlari |
| **errorcode** | supporting | `admin/supporting/errorcode` | Xato kodlari katalogi |

---

## Shared Kernel (`internal/kernel/`)

```
internal/kernel/
├── domain/                        # Umumiy value objectlar & konstantalar
│   └── consts/
│
├── application/                   # EventBus interfeysi, CQRS base typelar
│
└── infrastructure/                # Cross-cutting tashkilotlar
    ├── db/
    │   ├── postgres/             # PostgreSQL ulanish pool (pgxpool) + Squirrel
    │   ├── redis/                # Redis ulanish
    │   └── minio/                # MinIO (S3-compatible) object storage
    ├── security/
    │   ├── jwt/                  # JWT (RSA) + device fingerprinting
    │   └── csrf/                 # CSRF (HMAC-SHA256, Redis-backed)
    ├── logger/                   # Zap structured logging
    ├── tracing/                  # OpenTelemetry + Jaeger
    ├── metrics/                  # OpenTelemetry + Prometheus
    ├── asynq/                    # Background job queue (Redis-backed)
    ├── cache/                    # In-memory cache (LRU, LFU, SLRU, 2Q, FIFO, ...)
    ├── eventbus/                 # In-Memory event bus
    ├── errors/                   # Typed xato kodlar, DB xato mapping, retry
    ├── errorx/                   # HTTP xato response helper'lar
    ├── httpx/                    # HTTP request/response utillar
    ├── middleware/                # CORS, security headers, rate limiting
    ├── sse/                      # Server-Sent Events
    ├── pubsub/                   # Pub/sub messaging
    ├── metadata/                 # Entity metadata (EAV key-value store)
    ├── validator/                # Struct validatsiya
    ├── validation/               # Field validatorlar (email, phone, password, UUID)
    ├── contextx/                 # Type-safe context kalitlar
    ├── pgxutil/                  # Transaction helper'lar (WithTx)
    ├── ptrutil/                  # Generic pointer utillar
    ├── useragent/                # User-Agent parser
    ├── server/http/              # HTTP server
    ├── container/                # Test container helper'lar (testcontainers-go)
    ├── firebase/                 # Firebase Cloud Messaging
    └── telegram/                 # Telegram bot
```

---

## Ishga Tushirish Oqimi

```
cmd/app/main.go
  └── config.NewConfig()
        └── app.Run(cfg)
              ├── 1. Logger (Zap)
              ├── 2. Telemetry (OpenTelemetry + Jaeger)
              ├── 3. PostgreSQL connection pool
              ├── 4. Redis connection
              ├── 5. Asynq task queue client
              ├── 6. DDD: NewDDDBoundedContexts()  ← barcha BC'larni yaratish
              ├── 7. EventBus (InMemory)
              ├── 8. Cache invalidation service
              ├── 9. Asynq Worker
              ├── 10. HTTP Router (Gin) + Middleware + Routes
              └── 11. Graceful Shutdown (SIGINT/SIGTERM)
```

**Bootstrap fayllari:**
- `internal/app/app.go` — asosiy Run() funksiyasi
- `internal/app/ddd_bootstrap.go` — barcha BC'larni yaratadi
- `internal/app/ddd_routes.go` — DDD HTTP routelarni ro'yxatga oladi

---

## Texnologiyalar Steki

| Komponent | Texnologiya |
|-----------|-------------|
| Til | Go 1.26 |
| Web Framework | Gin |
| Database | PostgreSQL 18 |
| Query Builder | Squirrel |
| Cache | Redis 7.4 (go-redis) |
| Task Queue | Asynq (Redis-backed) |
| File Storage | MinIO (S3-compatible) |
| Logging | Zap (structured) |
| Metrics | OpenTelemetry + Prometheus |
| Tracing | OpenTelemetry + Jaeger |
| Validation | validator/v10 |
| API Docs | Swagger (swaggo/swag) |
| Testing | testify, testcontainers |
| Migrations | Goose v3 |
| Auth | JWT (RSA) + CSRF (Double-Submit Cookie) |

---

## Asosiy DDD Patternlar

### CQRS (Command Query Responsibility Segregation)
```
Command (write) → CommandHandler → WriteRepository → DB
                                 → EventBus.Publish(DomainEvent)

Query (read)    → QueryHandler   → ReadRepository  → DB
```

### Domain Event Flow
```
CommandHandler yaratadi → EventBus publish qiladi → Subscribers react qiladi
```

### Repository Pattern
```go
// domain/repository.go (Port — interfeys)
type UserWriteRepository interface {
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
}

// infrastructure/postgres/write_repo.go (Adapter — implementatsiya)
type UserWriteRepo struct { db *postgres.Postgres }
func (r *UserWriteRepo) Create(ctx context.Context, user *User) error { ... }
```

### Bounded Context Wiring
```go
// bc.go — har bir BC o'zini wire qiladi
type BoundedContext struct {
    CreateUser  *command.CreateUserHandler
    UpdateUser  *command.UpdateUserHandler
    GetUser     *query.GetUserHandler
    ListUsers   *query.ListUsersHandler
}
```
