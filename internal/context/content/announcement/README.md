# Announcement

Bounded context for managing multilingual announcements with publish workflow, priority ordering, and optional scheduling via start/end dates.

## Domain

### Aggregate Root
- `Announcement` -- Multilingual announcement with fields: `title` (Lang), `content` (Lang), `published` (bool), `publishedAt`, `priority` (int), `startDate`, and `endDate`. Title and content use the shared `Lang` value object supporting Uzbek, Russian, and English translations.

### Value Objects
- `shared.Lang` -- Trilingual text container with `Uz`, `Ru`, `En` fields (defined in the shared domain).

### Domain Events
- `AnnouncementPublished` -- Raised when an announcement transitions to published state via the `Publish()` method.

### Domain Errors
- `ErrAnnouncementNotFound` -- Returned when a requested announcement does not exist.
- `ErrAlreadyPublished` -- Returned when attempting to publish an already-published announcement.

### Repository Interfaces
- `AnnouncementRepository` (write) -- `Save`, `FindByID`, `Update`, `Delete`, `List`
- `AnnouncementReadRepository` (read) -- `FindByID`, `List` (returns `AnnouncementView` projections with flattened language columns)

### Filter
- `AnnouncementFilter` -- Supports filtering by `Published` (bool) and `Priority` (int), with `Limit`/`Offset` pagination.

## Application (CQRS)

### Commands
- `CreateAnnouncementCommand` / `CreateAnnouncementHandler` -- Creates a new unpublished announcement with multilingual title/content, priority, and optional date range.
- `UpdateAnnouncementCommand` / `UpdateAnnouncementHandler` -- Partially updates announcement fields. Supports an optional `Publish` flag that transitions the announcement to published state if not already published.
- `DeleteAnnouncementCommand` / `DeleteAnnouncementHandler` -- Hard-deletes an announcement by ID.

### Queries
- `GetAnnouncementQuery` / `GetAnnouncementHandler` -- Fetches a single announcement view by ID, reconstructing the `Lang` value objects from flattened DB columns.
- `ListAnnouncementsQuery` / `ListAnnouncementsHandler` -- Returns a paginated list of announcement views with total count.

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/announcements` | Create a new announcement |
| GET | `/announcements` | List announcements (paginated via `limit`/`offset` query params) |
| GET | `/announcements/:id` | Get a single announcement by ID |
| PATCH | `/announcements/:id` | Partially update an announcement (set `publish: true` to publish) |
| DELETE | `/announcements/:id` | Delete an announcement |

## Usage
```go
import "gct/internal/announcement"
```
