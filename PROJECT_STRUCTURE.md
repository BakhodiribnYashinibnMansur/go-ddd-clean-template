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
│   ├── shared/                   # Shared Kernel
│   └── [19 Bounded Context]      # DDD domenlar
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

Har bir Bounded Context quyidagi hexagonal/onion arxitektura layerlaridan iborat:

```
internal/{bc_name}/
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

## 19 Bounded Context ro'yxati

### Core Domenlar
| BC | Vazifasi |
|----|----------|
| **user** | Foydalanuvchi boshqaruvi, autentifikatsiya (SignIn, SignUp, CRUD) |
| **authz** | Avtorizatsiya — RBAC (Role, Permission, Policy, Scope) |
| **session** | Sessiya boshqaruvi, device tracking |
| **audit** | Audit logging |

### Feature Domenlar
| BC | Vazifasi |
|----|----------|
| **announcement** | Tizim e'lonlari (ko'p tilli) |
| **dashboard** | Dashboard ma'lumotlari agregatsiyasi |
| **dataexport** | Ma'lumot eksport ishlari |
| **errorcode** | Xato kodlari boshqaruvi |
| **featureflag** | Feature toggle |
| **file** | Fayl metadata boshqaruvi (MinIO) |
| **integration** | Tashqi tizimlar integratsiyasi, API kalitlar |
| **iprule** | IP asosida kirish nazorati |
| **metric** | Performance metrikalari |
| **notification** | Foydalanuvchi bildirishnomalari |
| **ratelimit** | Rate limiting qoidalari |
| **sitesetting** | Sayt konfiguratsiyasi |
| **systemerror** | Tizim xatolarini kuzatish (5xx) |
| **translation** | Ko'p tilli qo'llab-quvvatlash |
| **usersetting** | Foydalanuvchi sozlamalari |

---

## Shared Kernel (`internal/shared/`)

```
internal/shared/
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
              ├── 6. DDD: NewDDDBoundedContexts()  ← 19 BC yaratish
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
| Til | Go 1.25 |
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
