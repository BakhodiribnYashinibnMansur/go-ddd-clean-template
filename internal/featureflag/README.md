# Feature Flag

Bounded context for feature flag management with rollout percentage support. Controls feature availability through enable/disable toggles and gradual rollout via configurable percentages.

## Domain

### Aggregate Root
- `FeatureFlag` -- Represents a feature flag. Key fields: `name`, `description`, `enabled`, `rolloutPercentage` (0-100 integer). Embeds `shared.AggregateRoot` (provides `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt`).
  - `Toggle()` -- Flips the `enabled` boolean and raises a `FlagToggled` event.
  - `UpdateDetails(name, description, enabled, rolloutPercentage)` -- Partial update of mutable fields (nil pointers are skipped).

### Domain Events
- `FlagToggled` -- Raised when `Toggle()` is called. Carries the aggregate ID and the new enabled state.

### Repository Interfaces
- `FeatureFlagRepository` (write) -- `Save`, `FindByID`, `Update`, `Delete`
- `FeatureFlagReadRepository` (read) -- `FindByID`, `List` (with `FeatureFlagFilter`: search, enabled, limit, offset)

## Application (CQRS)

### Commands
- `CreateCommand` / `CreateHandler` -- Creates a new feature flag with name, description, enabled state, and rollout percentage. Publishes domain events.
- `UpdateCommand` / `UpdateHandler` -- Loads a feature flag by ID, applies partial field updates, persists, and publishes events.
- `DeleteCommand` / `DeleteHandler` -- Deletes a feature flag by ID.

### Queries
- `GetQuery` / `GetHandler` -- Fetches a single feature flag by ID and returns a `FeatureFlagView` DTO.
- `ListQuery` / `ListHandler` -- Returns a paginated, filtered list of feature flags with total count.

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /feature-flags | Create a new feature flag |
| GET | /feature-flags | List feature flags (paginated via `limit`/`offset` query params) |
| GET | /feature-flags/:id | Get a single feature flag by ID |
| PATCH | /feature-flags/:id | Partially update a feature flag |
| DELETE | /feature-flags/:id | Delete a feature flag |

## Usage
```go
import "gct/internal/featureflag"
```
