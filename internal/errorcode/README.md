# Error Code

Manages application-wide error code definitions with rich metadata including HTTP status, category, severity, retry policy, and user-facing suggestions. Provides a centralized registry so error responses are consistent and configurable without code changes.

## Domain

### Aggregate Root
- `ErrorCode` -- Represents a single error code entry. Key fields: `code` (unique string identifier), `message`, `httpStatus`, `category`, `severity`, `retryable` (bool), `retryAfter` (seconds), `suggestion` (user-facing hint). Embeds `shared.AggregateRoot`.

### Domain Events
- `ErrorCodeUpdated` -- Raised when an error code is created or updated. Carries the aggregate ID, `Code`, and `Message`.

### Domain Errors
- `ErrErrorCodeNotFound` -- Returned when an error code cannot be found by ID.

### Repository Interfaces
- `ErrorCodeRepository` (write) -- `Save`, `Update`, `FindByID`, `Delete`
- `ErrorCodeReadRepository` (read) -- `FindByID`, `List` (returns `ErrorCodeView` projections)

### Filter
- `ErrorCodeFilter` -- Supports filtering by `Code`, `Category`, `Severity`, plus `Limit`/`Offset` pagination.

## Application (CQRS)

### Commands
- `CreateErrorCodeCommand` / `CreateErrorCodeHandler` -- Creates a new error code with all metadata fields and publishes the `ErrorCodeUpdated` event.
- `UpdateErrorCodeCommand` / `UpdateErrorCodeHandler` -- Fully replaces all mutable fields of an existing error code (all fields except `code` are required). The `code` field is immutable after creation. Publishes the `ErrorCodeUpdated` event.
- `DeleteErrorCodeCommand` / `DeleteErrorCodeHandler` -- Deletes an error code by ID.

### Queries
- `GetErrorCodeQuery` / `GetErrorCodeHandler` -- Fetches a single error code view by ID.
- `ListErrorCodesQuery` / `ListErrorCodesHandler` -- Returns a paginated, filtered list of error code views with total count.

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/error-codes` | Create a new error code |
| GET | `/error-codes` | List error codes (paginated: `limit`, `offset` query params) |
| GET | `/error-codes/:id` | Get a single error code by ID |
| PATCH | `/error-codes/:id` | Update an error code (full replacement of mutable fields) |
| DELETE | `/error-codes/:id` | Delete an error code |

## Usage
```go
import "gct/internal/errorcode"
```
