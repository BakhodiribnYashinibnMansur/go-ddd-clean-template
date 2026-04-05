# Integration

Bounded context for managing third-party integrations. Stores connection details including API keys, webhook URLs, integration type, and arbitrary configuration for external service connectors.

## Domain

### Aggregate Root
- `Integration` -- Represents a third-party integration. Key fields: `name`, `intType` (integration type, e.g. "slack", "stripe"), `apiKey`, `webhookURL`, `enabled`, `config` (arbitrary JSON configuration map). Embeds `shared.AggregateRoot` (provides `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt`).
  - `UpdateDetails(name, intType, apiKey, webhookURL, enabled, config)` -- Partial update of mutable fields (nil pointers are skipped).

### Domain Events
- `IntegrationConnected` -- Raised when a new integration is created via `NewIntegration()`. Carries the aggregate ID, name, and type.

### Repository Interfaces
- `IntegrationRepository` (write) -- `Save`, `FindByID`, `Update`, `Delete`
- `IntegrationReadRepository` (read) -- `FindByID`, `List` (with `IntegrationFilter`: search, type, enabled, limit, offset)

## Application (CQRS)

### Commands
- `CreateCommand` / `CreateHandler` -- Creates a new integration with name, type, API key, webhook URL, enabled state, and config. Publishes `IntegrationConnected` event.
- `UpdateCommand` / `UpdateHandler` -- Loads an integration by ID, applies partial field updates, persists, and publishes events.
- `DeleteCommand` / `DeleteHandler` -- Deletes an integration by ID.

### Queries
- `GetQuery` / `GetHandler` -- Fetches a single integration by ID and returns an `IntegrationView` DTO.
- `ListQuery` / `ListHandler` -- Returns a paginated, filtered list of integrations with total count.

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /integrations | Create a new integration |
| GET | /integrations | List integrations (paginated via `limit`/`offset` query params) |
| GET | /integrations/:id | Get a single integration by ID |
| PATCH | /integrations/:id | Partially update an integration |
| DELETE | /integrations/:id | Delete an integration |

## Usage
```go
import "gct/internal/integration"
```
