# System Error

Bounded context for recording, tracking, and resolving system errors with full stack traces and request context metadata.

## Domain

### Aggregate Root
- `SystemError` -- Represents a captured system error. Key fields: `code`, `message`, `stackTrace`, `metadata`, `severity`, `serviceName`, `requestID`, `userID`, `ipAddress`, `path`, `method`, `isResolved`, `resolvedAt`, `resolvedBy`. Embeds `shared.AggregateRoot`.

### Domain Events
- `SystemErrorRecorded` -- Raised when a new system error is recorded via `NewSystemError()`.
- `SystemErrorResolved` -- Raised when `Resolve(resolvedBy)` is called on an existing error.

### Domain Errors
- `ErrSystemErrorNotFound` -- Returned when a system error cannot be found by ID.

### Repository Interfaces
- `SystemErrorRepository` (write) -- `Save`, `FindByID`, `Update`, `List`
- `SystemErrorReadRepository` (read) -- `FindByID`, `List` (returns `SystemErrorView` projections)

### Filter
- `SystemErrorFilter` -- Supports filtering by `Code`, `Severity`, `IsResolved`, `FromDate`, `ToDate`, `RequestID`, `UserID`, with `Limit`/`Offset` pagination.

## Application (CQRS)

### Commands
- `CreateSystemErrorCommand` / `CreateSystemErrorHandler` -- Records a new system error with optional stack trace, metadata, and request context. Publishes `SystemErrorRecorded` event.
- `ResolveErrorCommand` / `ResolveErrorHandler` -- Marks an existing system error as resolved by a given user. Publishes `SystemErrorResolved` event.

### Queries
- `GetSystemErrorQuery` / `GetSystemErrorHandler` -- Fetches a single system error by ID, returns `SystemErrorView`.
- `ListSystemErrorsQuery` / `ListSystemErrorsHandler` -- Lists system errors with filtering, returns `[]*SystemErrorView` and total count.

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/system-errors` | Record a new system error |
| GET | `/system-errors` | List system errors (paginated via `limit`/`offset` query params) |
| GET | `/system-errors/:id` | Get a single system error by ID |
| POST | `/system-errors/:id/resolve` | Mark a system error as resolved |

## Usage
```go
import "gct/internal/systemerror"
```
