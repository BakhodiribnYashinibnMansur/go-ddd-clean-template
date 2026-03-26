# User Setting

Bounded context for managing per-user key-value settings with upsert semantics.

## Domain

### Aggregate Root
- `UserSetting` -- Represents a single user setting as a key-value pair. Key fields: `userID`, `key`, `value`. Embeds `shared.AggregateRoot`. Supports `ChangeValue()` to update the value in place.

### Domain Events
- `UserSettingChanged` -- Raised when a user setting is created or its value is updated.

### Domain Errors
- `ErrUserSettingNotFound` -- Returned when a user setting cannot be found.

### Repository Interfaces
- `UserSettingRepository` (write) -- `Upsert`, `FindByUserIDAndKey`, `Delete`
- `UserSettingReadRepository` (read) -- `FindByID`, `List` (returns `UserSettingView` projections)

### Filter
- `UserSettingFilter` -- Supports filtering by `UserID`, `Key`, with `Limit`/`Offset` pagination.

## Application (CQRS)

### Commands
- `UpsertUserSettingCommand` / `UpsertUserSettingHandler` -- Creates a new setting or updates an existing one (matched by `userID` + `key`). Publishes `UserSettingChanged` event.
- `DeleteUserSettingCommand` / `DeleteUserSettingHandler` -- Deletes a user setting by ID.

### Queries
- `GetUserSettingQuery` / `GetUserSettingHandler` -- Fetches a single user setting by ID, returns `UserSettingView`.
- `ListUserSettingsQuery` / `ListUserSettingsHandler` -- Lists user settings with filtering, returns `[]*UserSettingView` and total count.

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/user-settings` | Create or update a user setting (upsert) |
| GET | `/user-settings` | List user settings (paginated via `limit`/`offset` query params) |
| DELETE | `/user-settings/:id` | Delete a user setting by ID |

## Usage
```go
import "gct/internal/usersetting"
```
