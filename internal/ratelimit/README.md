# Rate Limit

Manages rate limiting rules that control how many requests are allowed within a given time window. Each rule targets a specific endpoint or pattern and can be enabled or disabled independently.

## Domain

### Aggregate Root
- `RateLimit` -- Defines a rate limiting rule with a name, matching rule pattern, requests-per-window cap, window duration (seconds), and enabled flag. Embeds `shared.AggregateRoot` (ID, CreatedAt, UpdatedAt).

### Domain Events
- `RateLimitChanged` -- Raised when a rate limit is created or updated. Carries the aggregate ID.

### Domain Errors
- `ErrRateLimitNotFound` -- Returned when a rate limit cannot be found by ID.

### Repository Interfaces
- `RateLimitRepository` (write) -- `Save`, `FindByID`, `Update`, `Delete`, `List`
- `RateLimitReadRepository` (read) -- `FindByID`, `List` (returns `RateLimitView` projections)

### Filter
- `RateLimitFilter` -- Supports filtering by `Name`, `Enabled`, plus `Limit`/`Offset` pagination.

## Application (CQRS)

### Commands
- `CreateRateLimitCommand` / `CreateRateLimitHandler` -- Creates a new rate limit rule and publishes domain events.
- `UpdateRateLimitCommand` / `UpdateRateLimitHandler` -- Partially updates an existing rate limit (all fields optional via pointers) and publishes domain events.
- `DeleteRateLimitCommand` / `DeleteRateLimitHandler` -- Deletes a rate limit by ID.

### Queries
- `GetRateLimitQuery` / `GetRateLimitHandler` -- Fetches a single rate limit view by ID.
- `ListRateLimitsQuery` / `ListRateLimitsHandler` -- Returns a paginated, filtered list of rate limit views with total count.

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/rate-limits` | Create a new rate limit rule |
| GET | `/rate-limits` | List rate limits (paginated: `limit`, `offset` query params) |
| GET | `/rate-limits/:id` | Get a single rate limit by ID |
| PATCH | `/rate-limits/:id` | Partially update a rate limit |
| DELETE | `/rate-limits/:id` | Delete a rate limit |

## Usage
```go
import "gct/internal/ratelimit"
```
