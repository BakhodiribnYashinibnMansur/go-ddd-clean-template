# Session

Read-only bounded context for querying session data. Sessions are created and managed by the User aggregate; this package provides a dedicated read-side projection for listing and inspecting sessions independently.

## Domain

This bounded context has no domain model of its own. Sessions are child entities of the User aggregate (defined in `internal/user/domain`). The session package only reads from the `sessions` table via a read-model DTO.

### Read-Model DTO
- `SessionView` -- ID, UserID, DeviceID, DeviceName, DeviceType, IPAddress, UserAgent, ExpiresAt, LastActivity, Revoked, CreatedAt
- `SessionsFilter` -- optional filters: UserID, Revoked, Limit, Offset

## Application (CQRS -- Query Side Only)

### Repository Interface
- `SessionReadRepository` (read) -- `FindByID(id)`, `List(filter)`

### Queries
- `GetSessionQuery` / `GetSessionHandler` -- fetches a single session by ID, returns `SessionView`
- `ListSessionsQuery` / `ListSessionsHandler` -- lists sessions with optional UserID/Revoked filters and pagination, returns `[]*SessionView` with total count

## Infrastructure

### PostgreSQL Read Repository
- `SessionReadRepo` -- implements `SessionReadRepository` using pgxpool and Squirrel query builder; reads from the `sessions` table with dollar-placeholder parameterized queries

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /sessions | List sessions with optional `user_id`, `limit`, `offset` query params |
| GET | /sessions/:id | Get a single session by UUID |

## Wiring

`BoundedContext` struct wires the two query handlers together. Created via `NewBoundedContext(pool, logger)`.

## Usage
```go
import "gct/internal/session"
```
