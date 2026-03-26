# Webhook

Bounded context for managing webhook endpoints with event subscriptions. Supports registering callback URLs that listen for specific event types, with secret-based signing and enable/disable control.

## Domain

### Aggregate Root
- `Webhook` -- Represents a webhook endpoint. Key fields: `name`, `url`, `secret`, `events` (subscribed event names), `enabled`. Embeds `shared.AggregateRoot` (provides `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt`).
  - `Trigger()` -- Adds a `WebhookTriggered` domain event.
  - `UpdateDetails(name, url, secret, events, enabled)` -- Partial update of mutable fields (nil pointers are skipped).

### Domain Events
- `WebhookTriggered` -- Raised when `Trigger()` is called on a webhook. Carries the aggregate ID and target URL.

### Repository Interfaces
- `WebhookRepository` (write) -- `Save`, `FindByID`, `Update`, `Delete`
- `WebhookReadRepository` (read) -- `FindByID`, `List` (with `WebhookFilter`: search, enabled, limit, offset)

## Application (CQRS)

### Commands
- `CreateCommand` / `CreateHandler` -- Creates a new `Webhook` aggregate and persists it. Publishes domain events via the event bus.
- `UpdateCommand` / `UpdateHandler` -- Loads a webhook by ID, applies partial field updates, persists, and publishes events.
- `DeleteCommand` / `DeleteHandler` -- Deletes a webhook by ID.

### Queries
- `GetQuery` / `GetHandler` -- Fetches a single webhook by ID and returns a `WebhookView` DTO.
- `ListQuery` / `ListHandler` -- Returns a paginated, filtered list of webhooks with total count.

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /webhooks | Create a new webhook |
| GET | /webhooks | List webhooks (paginated via `limit`/`offset` query params) |
| GET | /webhooks/:id | Get a single webhook by ID |
| PATCH | /webhooks/:id | Partially update a webhook |
| DELETE | /webhooks/:id | Delete a webhook |

## Usage
```go
import "gct/internal/webhook"
```
