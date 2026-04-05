# Statistics

Read-only bounded context that aggregates system-wide counts across other bounded contexts for the admin panel. Has no domain layer of its own — it queries platform tables directly.

## Domain

No aggregate root, value objects, or domain events. Pure query-side (read-only) context.

### Repository Interface
- `StatisticsReadRepository` — one method per endpoint (`GetOverview`, `GetUserStats`, `GetSessionStats`, `GetErrorStats`, `GetAuditStats`, `GetSecurityStats`, `GetFeatureFlagStats`, `GetContentStats`, `GetIntegrationStats`).

## Application (CQRS)

### Commands
None. Read-only bounded context.

### Queries
| Query | Returns |
|-------|---------|
| `GetOverviewQuery` | `OverviewView` (`TotalUsers`, `ActiveSessions`, `AuditLogsToday`, `SystemErrorsCount`, `TotalFeatureFlags`) |
| `GetUserStatsQuery` | `UserStatsView` (`Total`, `Deleted`, `ByRole`) |
| `GetSessionStatsQuery` | `SessionStatsView` (`Active`, `Expired`, `Revoked`) |
| `GetErrorStatsQuery` | `ErrorStatsView` (`Unresolved`, `Resolved`, `Last24h`) |
| `GetAuditStatsQuery` | `AuditStatsView` (`Today`, `Last7Days`, `Total`) |
| `GetSecurityStatsQuery` | `SecurityStatsView` (`IPRules`, `RateLimits`) |
| `GetFeatureFlagStatsQuery` | `FeatureFlagStatsView` (`Total`, `Enabled`, `Disabled`) |
| `GetContentStatsQuery` | `ContentStatsView` (`Announcements`, `Notifications`, `FileMetadata`, `Translations`) |
| `GetIntegrationStatsQuery` | `IntegrationStatsView` (`Integrations`, `APIKeys`) |

## HTTP API

All endpoints are `GET` and wrapped as `{"data": …}`.

| Endpoint | Description |
|----------|-------------|
| `/statistics/overview` | Top-level aggregated counts |
| `/statistics/users` | User total, deleted, and per-role breakdown |
| `/statistics/sessions` | Active, expired, revoked sessions |
| `/statistics/errors` | Unresolved, resolved, last 24h system errors |
| `/statistics/audit` | Today, last 7 days, total audit log entries |
| `/statistics/security` | IP rules and rate limits counts |
| `/statistics/feature-flags` | Total, enabled, disabled feature flags |
| `/statistics/content` | Announcements, notifications, file metadata, translations |
| `/statistics/integrations` | Active integrations and API keys |

## Usage
```go
import "gct/internal/contexts/admin/statistics"
```
