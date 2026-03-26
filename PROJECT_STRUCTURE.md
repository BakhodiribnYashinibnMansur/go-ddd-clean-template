# Golang Template — DDD Architecture

## Umumiy Ko'rinish

```
Backend/
├── cmd/                           # Dastur kirish nuqtalari
│   ├── app/main.go               # Asosiy dastur
│   ├── seeder/                   # DB seed qilish
│   ├── keygen/                   # Kalit generatsiya
│   └── migration/                # Migratsiya ishga tushirish
│
├── config/                        # Konfiguratsiya
│   ├── config.go                 # Config struct & NewConfig()
│   ├── config.yaml               # YAML konfiguratsiya
│   ├── database.go               # DB sozlamalari
│   ├── jwt.go                    # JWT sozlamalari
│   └── redis.go                  # Redis sozlamalari
│
├── migrations/postgres/           # Goose SQL migratsiyalar (~32 fayl)
│
├── internal/                      # Asosiy business logika
│   ├── app/                      # Bootstrap & Wiring
│   ├── shared/                   # Shared Kernel
│   ├── [22 Bounded Context]      # DDD domenlar
│   ├── domain/                   # (Legacy) domen modellari
│   ├── usecase/                  # (Legacy) use case'lar
│   ├── repo/                     # (Legacy) repository'lar
│   └── controller/               # (Legacy) HTTP handlerlar
│
├── test/                          # Testlar
├── docs/                          # Swagger/OpenAPI
├── nginx/                         # Nginx config
├── Dockerfile
├── docker-compose.yml
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

## 22 Bounded Context ro'yxati

### Core Domenlar
| BC | Vazifasi |
|----|----------|
| **user** | Foydalanuvchi boshqaruvi, autentifikatsiya (SignIn, SignUp, CRUD) |
| **authz** | Avtorizatsiya — RBAC (Role, Permission, Policy, Scope) |
| **session** | Sessiya boshqaruvi |
| **audit** | Audit logging |

### Feature Domenlar
| BC | Vazifasi |
|----|----------|
| **announcement** | Tizim e'lonlari |
| **dashboard** | Dashboard ma'lumotlari agregatsiyasi |
| **featureflag** | Feature toggle |
| **integration** | Tashqi tizimlar integratsiyasi |
| **notification** | Foydalanuvchi bildirishnomalari |
| **translation** | Ko'p tilli qo'llab-quvvatlash |
| **file** | Fayl metadata boshqaruvi |
| **webhook** | Webhook boshqaruvi va yetkazish |
| **emailtemplate** | Email shablon saqlash |
| **dataexport** | Ma'lumot eksport ishlari |
| **sitesetting** | Sayt konfiguratsiyasi |
| **usersetting** | Foydalanuvchi sozlamalari |
| **ratelimit** | Rate limiting qoidalari |
| **iprule** | IP asosida kirish nazorati |
| **job** | Background job rejalashtirish |
| **systemerror** | Tizim xatolarini kuzatish |
| **errorcode** | Xato kodlari boshqaruvi |
| **metric** | Performance metrikalari |

---

## Shared Kernel (`internal/shared/`)

```
internal/shared/
├── domain/                        # Umumiy value objectlar & konstantalar
│   └── consts/
│
├── application/                   # Umumiy interfeyslar
│   └── event_bus.go              # EventBus interfeysi
│
└── infrastructure/                # Cross-cutting tashkilotlar
    ├── db/
    │   ├── postgres/             # PostgreSQL ulanish pool
    │   └── redis/                # Redis ulanish
    ├── logger/                   # Zap structured logging
    ├── security/                 # JWT, CSRF
    ├── server/http/              # HTTP server wrapper
    ├── cache/                    # Cache utillar
    ├── errorx/                   # Xato ishlov berish
    ├── asynq/                    # Task queue (Redis-backed)
    ├── validator/                # Input validatsiya
    ├── tracing/                  # OpenTelemetry
    ├── eventbus/                 # In-Memory event bus
    ├── httpx/                    # HTTP utillar
    ├── contextx/                 # Context utillar
    ├── pgxutil/                  # PostgreSQL utillar
    ├── container/                # DI container
    ├── firebase/                 # Firebase integratsiya
    ├── telegram/                 # Telegram bot
    └── validation/               # Validatsiya qoidalari
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
              ├── 6. Legacy: Repositories → UseCases
              ├── 7. DDD: NewDDDBoundedContexts()  ← 22 BC yaratish
              ├── 8. EventBus (InMemory)
              ├── 9. Cache invalidation service
              ├── 10. Asynq Worker
              ├── 11. HTTP Router (Gin) + Middleware + Routes
              └── 12. Graceful Shutdown (SIGINT/SIGTERM)
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
| Database | PostgreSQL (NeonDB) |
| Query Builder | Squirrel |
| Code Gen | sqlc |
| Cache | Redis (go-redis) |
| Task Queue | Asynq |
| File Storage | MinIO (S3-compatible) |
| Logging | Zap |
| Tracing | OpenTelemetry + Jaeger |
| Validation | validator/v10 |
| API Docs | Swagger (swaggo/swag) |
| Testing | testify, miniredis |
| Migrations | Goose |
| Auth | JWT + CSRF (Double-Submit Cookie) |

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

---

## Legacy vs DDD

| Layer | Legacy (o'chiriladi) | DDD (yangi) |
|-------|---------------------|-------------|
| Entity | `internal/domain/` | `internal/{bc}/domain/` |
| UseCase | `internal/usecase/` | `internal/{bc}/application/command\|query/` |
| Repo | `internal/repo/` | `internal/{bc}/infrastructure/postgres/` |
| Handler | `internal/controller/` | `internal/{bc}/interfaces/http/` |

> Legacy kod hali ham mavjud va bosqichma-bosqich DDD'ga ko'chirilmoqda.

---

## Migratsiya Holati

**Bajarilgan:**
- Shared Kernel (BaseEntity, AggregateRoot, Repository[T], CommandHandler, QueryHandler)
- 22 Bounded Context (domain + application + infrastructure layerlar)
- InMemoryEventBus + KafkaEventBus stub
- DDDBoundedContexts container

**Qolgan ishlar:**
- HTTP interfaces layer (handlerlar + routelar)
- Kafka EventBus to'liq implementatsiya
- Cross-BC event subscribers wiring
- Legacy kodlarni o'chirish
- Integration/E2E testlar
