# Audit

Bounded context for immutable audit logging and HTTP endpoint history tracking. Supports a wide range of auditable actions across authentication, authorization, user management, and business operations.

## Domain

### Aggregate Root
- `AuditLog` -- Represents an immutable audit log entry. Key fields: `userID`, `sessionID`, `action` (AuditAction enum), `resourceType`, `resourceID`, `platform`, `ipAddress`, `userAgent`, `permission`, `policyID`, `decision`, `success`, `errorMessage`, `metadata`. Embeds `shared.AggregateRoot`.

### Value Objects
- `AuditAction` -- String enum representing the type of auditable action. Values include: `LOGIN`, `LOGOUT`, `SESSION_REVOKE`, `PASSWORD_CHANGE`, `MFA_VERIFY_FAIL`, `ACCESS_GRANTED`, `ACCESS_DENIED`, `POLICY_MATCHED`, `POLICY_DENIED`, `USER_CREATE`, `USER_UPDATE`, `USER_DELETE`, `ROLE_ASSIGN`, `ROLE_REMOVE`, `ORDER_APPROVE`, `ORDER_CANCEL`, `PAYMENT_PROCESS`, `PAYMENT_CANCEL`, `POLICY_EVALUATED`, `ADMIN_CHANGE`.

### Entities
- `EndpointHistory` -- Tracks HTTP request history. Key fields: `userID`, `endpoint`, `method`, `statusCode`, `latency`, `ipAddress`, `userAgent`. Embeds `shared.BaseEntity` (not an aggregate root).

### Domain Events
- `AuditLogCreated` -- Raised when a new audit log entry is created.

### Domain Errors
- `ErrAuditLogNotFound` -- Returned when an audit log cannot be found.

### Repository Interfaces
- `AuditLogRepository` (write) -- `Save` (immutable, append-only)
- `EndpointHistoryRepository` (write) -- `Save` (immutable, append-only)
- `AuditReadRepository` (read) -- `ListAuditLogs`, `ListEndpointHistory` (returns view projections)

### Filters
- `AuditLogFilter` -- Supports filtering by `UserID`, `Action`, `ResourceType`, `ResourceID`, `Success`, `FromDate`, `ToDate`, with `Pagination`.
- `EndpointHistoryFilter` -- Supports filtering by `UserID`, `Method`, `Endpoint`, `StatusCode`, `FromDate`, `ToDate`, with `Pagination`.

## Application (CQRS)

### Commands
- `CreateAuditLogCommand` / `CreateAuditLogHandler` -- Creates a new immutable audit log entry. Publishes `AuditLogCreated` event.
- `CreateEndpointHistoryCommand` / `CreateEndpointHistoryHandler` -- Creates a new immutable endpoint history entry. No events published.

### Queries
- `ListAuditLogsQuery` / `ListAuditLogsHandler` -- Lists audit logs with filtering, returns `[]*AuditLogView` and total count.
- `ListEndpointHistoryQuery` / `ListEndpointHistoryHandler` -- Lists endpoint history with filtering, returns `[]*EndpointHistoryView` and total count.

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/audit-logs` | List audit log entries (paginated via `limit`/`offset` query params) |
| GET | `/endpoint-history` | List endpoint history entries (paginated via `limit`/`offset` query params) |

## Usage
```go
import "gct/internal/audit"
```
