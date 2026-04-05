# IP Rule

Bounded context for managing IP-based access rules. Each rule targets a specific IP address with an ALLOW or DENY action and an optional expiry timestamp.

## Domain

### Aggregate Root
- `IPRule` -- Represents an IP access rule with fields: `ipAddress`, `action` (ALLOW/DENY), `reason`, and optional `expiresAt`. Embeds `shared.AggregateRoot` for identity, timestamps, and event tracking.

### Domain Events
- `IPRuleCreated` -- Raised when a new IP rule is created. Carries the aggregate ID, IP address, and action.

### Domain Errors
- `ErrIPRuleNotFound` -- Returned when a requested IP rule does not exist.

### Repository Interfaces
- `IPRuleRepository` (write) -- `Save`, `FindByID`, `Update`, `Delete`, `List`
- `IPRuleReadRepository` (read) -- `FindByID`, `List` (returns `IPRuleView` projections)

### Filter
- `IPRuleFilter` -- Supports filtering by `IPAddress`, `Action`, with `Limit`/`Offset` pagination.

## Application (CQRS)

### Commands
- `CreateIPRuleCommand` / `CreateIPRuleHandler` -- Creates a new IP rule aggregate, persists it, and publishes domain events.
- `UpdateIPRuleCommand` / `UpdateIPRuleHandler` -- Loads an existing IP rule by ID, applies partial updates (ipAddress, action, reason, expiresAt), persists, and publishes events.
- `DeleteIPRuleCommand` / `DeleteIPRuleHandler` -- Hard-deletes an IP rule by ID. No soft-delete or audit trail at this level.

### Queries
- `GetIPRuleQuery` / `GetIPRuleHandler` -- Fetches a single IP rule view by ID from the read repository.
- `ListIPRulesQuery` / `ListIPRulesHandler` -- Returns a paginated list of IP rule views with total count.

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/ip-rules` | Create a new IP rule |
| GET | `/ip-rules` | List IP rules (paginated via `limit`/`offset` query params) |
| GET | `/ip-rules/:id` | Get a single IP rule by ID |
| PATCH | `/ip-rules/:id` | Partially update an IP rule |
| DELETE | `/ip-rules/:id` | Delete an IP rule |

## Usage
```go
import "gct/internal/iprule"
```
