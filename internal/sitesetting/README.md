# Site Setting

Manages global site configuration as key-value pairs with typed values and descriptions. Provides a flexible way to store and retrieve application-wide settings (e.g., site name, feature flags, limits) without redeployment.

## Domain

### Aggregate Root
- `SiteSetting` -- Represents a single configuration entry. Key fields: `key` (setting identifier), `value` (string representation), `settingType` (data type hint, e.g., "string", "int", "bool"), `description` (human-readable explanation). Embeds `shared.AggregateRoot`.

### Domain Events
- `SettingUpdated` -- Raised when a site setting is updated. Carries the aggregate ID.

### Domain Errors
- `ErrSiteSettingNotFound` -- Returned when a site setting cannot be found by ID.

### Repository Interfaces
- `SiteSettingRepository` (write) -- `Save`, `FindByID`, `Update`, `Delete`, `List`
- `SiteSettingReadRepository` (read) -- `FindByID`, `List` (returns `SiteSettingView` projections)

### Filter
- `SiteSettingFilter` -- Supports filtering by `Key` and `Type`, plus `Limit`/`Offset` pagination.

## Application (CQRS)

### Commands
- `CreateSiteSettingCommand` / `CreateSiteSettingHandler` -- Creates a new site setting with key, value, type, and description.
- `UpdateSiteSettingCommand` / `UpdateSiteSettingHandler` -- Partially updates an existing site setting (all fields optional via pointers). Raises `SettingUpdated` event.
- `DeleteSiteSettingCommand` / `DeleteSiteSettingHandler` -- Deletes a site setting by ID.

### Queries
- `GetSiteSettingQuery` / `GetSiteSettingHandler` -- Fetches a single site setting view by ID.
- `ListSiteSettingsQuery` / `ListSiteSettingsHandler` -- Returns a paginated, filtered list of site setting views with total count.

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/site-settings` | Create a new site setting |
| GET | `/site-settings` | List site settings (paginated: `limit`, `offset` query params) |
| GET | `/site-settings/:id` | Get a single site setting by ID |
| PATCH | `/site-settings/:id` | Partially update a site setting |
| DELETE | `/site-settings/:id` | Delete a site setting |

## Usage
```go
import "gct/internal/sitesetting"
```
