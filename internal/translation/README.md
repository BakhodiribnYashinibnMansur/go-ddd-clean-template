# Translation

Bounded context for internationalization (i18n) translation management. Stores key-value translation pairs organized by language and logical group.

## Domain

### Aggregate Root
- `Translation` -- Represents a single translation entry. Key fields: `key` (lookup identifier), `language` (locale code), `value` (translated text), `group` (logical grouping). Embeds `shared.AggregateRoot`.
  - `Update(key, language, value, group)` -- Partial update of mutable fields (nil pointers are skipped). Raises `TranslationUpdated` event.

### Domain Events
- `TranslationUpdated` -- Raised when a translation is modified via `Update()`. Carries the aggregate ID.

### Repository Interfaces
- `TranslationRepository` (write) -- `Save`, `FindByID`, `Update`, `Delete`, `List`
- `TranslationReadRepository` (read) -- `FindByID`, `List` (with `TranslationFilter`: key, language, group, limit, offset)

## Application (CQRS)

### Commands
- `CreateTranslationCommand` / `CreateTranslationHandler` -- Creates a new translation with key, language, value, and group. Publishes domain events.
- `UpdateTranslationCommand` / `UpdateTranslationHandler` -- Loads a translation by ID, applies partial updates, persists, and publishes events.
- `DeleteTranslationCommand` / `DeleteTranslationHandler` -- Deletes a translation by ID.

### Queries
- `GetTranslationQuery` / `GetTranslationHandler` -- Fetches a single translation by ID and returns a `TranslationView` DTO.
- `ListTranslationsQuery` / `ListTranslationsHandler` -- Returns a paginated, filtered list of translations with total count.

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /translations | Create a new translation |
| GET | /translations | List translations (paginated via `limit`/`offset` query params) |
| GET | /translations/:id | Get a single translation by ID |
| PATCH | /translations/:id | Partially update a translation |
| DELETE | /translations/:id | Delete a translation |

## Usage
```go
import "gct/internal/translation"
```
