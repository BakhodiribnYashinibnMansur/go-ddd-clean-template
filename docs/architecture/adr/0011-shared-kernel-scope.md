# ADR-0011: Shared Kernel Scope

**Status:** ACCEPTED  
**Date:** 2026-04-07  

## Context

Bounded contexts need shared infrastructure: database connections, logging, caching,
event bus, HTTP middleware, validation, metrics, and security utilities. Duplicating
these in every BC would create maintenance burden and inconsistency. However, a
shared layer that grows unchecked becomes a "junk drawer" that couples all BCs to
implementation details.

## Decision

Maintain a shared kernel at `internal/kernel/` with clearly scoped sub-packages:

| Package                        | Responsibility                                |
|--------------------------------|-----------------------------------------------|
| `domain/`                      | Base domain event interface, aggregate root    |
| `application/`                 | EventBus, EventHandler interfaces              |
| `infrastructure/db/`           | PostgreSQL connection pool, pgx helpers        |
| `infrastructure/cache/`        | Redis client wrapper                           |
| `infrastructure/eventbus/`     | InMemory and Redis Streams event bus adapters  |
| `infrastructure/security/`     | JWT, keyring, CSRF, rate-limit, audit, TBH     |
| `infrastructure/middleware/`   | Gin middleware (auth, CORS, request logging)   |
| `infrastructure/logger/`       | Structured logging (slog-based)                |
| `infrastructure/validation/`   | Request validation helpers                     |
| `infrastructure/metrics/`      | Prometheus metric registration                 |
| `infrastructure/server/`       | HTTP/gRPC server lifecycle                     |
| `outbox/`                      | Transactional outbox entry, store, relay       |
| `consts/`                      | Shared constants                               |

The kernel exposes **interfaces** (ports) that BCs depend on, and
**implementations** (adapters) that are wired at application startup in `cmd/app`.
BCs import kernel interfaces but never import each other.

## Consequences

### Positive
- Single place to update infrastructure concerns (e.g., switch logging library).
- BCs stay lightweight -- no duplicated database or cache boilerplate.
- Kernel interfaces enforce a consistent contract across all BCs.

### Negative
- Changes to kernel interfaces require updating all BCs that depend on them.
- Risk of scope creep -- business logic must never enter the kernel.
- Tight coupling to kernel means all BCs upgrade infrastructure in lockstep.

## Alternatives Considered

- **Per-BC infrastructure** -- each BC owns its own DB connection and logger; leads
  to config drift, duplicated boilerplate, and inconsistent observability.
- **External Go packages** (separate module) -- adds versioning overhead and slows
  iteration; premature for a monorepo with a single deploy target.
- **No shared code** -- copy-paste of infrastructure utilities across 16 BCs;
  unmaintainable.
