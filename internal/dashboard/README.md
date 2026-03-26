# Dashboard

Read-only bounded context that aggregates statistics from across the system for dashboard display. Has no domain layer of its own -- it queries other bounded contexts' data directly.

## Domain

This bounded context has no aggregate root, value objects, or domain events. It is a pure query-side (read-only) context.

### Repository Interfaces
- `DashboardReadRepository` (read) -- `GetStats` (returns `DashboardStatsView`)

## Application (CQRS)

### Commands
None. This is a read-only bounded context.

### Queries
- `GetStatsQuery` / `GetStatsHandler` -- Fetches aggregated dashboard statistics including total users, active sessions, audit logs today, system errors count, total feature flags, total webhooks, and total jobs. Returns `DashboardStatsView`.

### DTO
- `DashboardStatsView` -- Fields: `TotalUsers`, `ActiveSessions`, `AuditLogsToday`, `SystemErrorsCount`, `TotalFeatureFlags`, `TotalWebhooks`, `TotalJobs`.

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/dashboard/stats` | Get aggregated dashboard statistics |

## Usage
```go
import "gct/internal/dashboard"
```
