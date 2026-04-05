# Notification

Bounded context for creating, listing, and managing user notifications with read-status tracking. Each notification is targeted to a specific user and supports marking as read via a `readAt` timestamp.

## Domain

### Aggregate Root
- `Notification` -- User-targeted notification with fields: `userID` (UUID), `title`, `message`, `nType` (notification type string), and optional `readAt` timestamp. Supports `MarkAsRead()` to set the read timestamp.

### Domain Events
- `NotificationSent` -- Raised when a new notification is created. Carries the aggregate ID, target user ID, and title.

### Domain Errors
- `ErrNotificationNotFound` -- Returned when a requested notification does not exist.

### Repository Interfaces
- `NotificationRepository` (write) -- `Save`, `FindByID`, `Update`, `Delete`
- `NotificationReadRepository` (read) -- `FindByID`, `List` (returns `NotificationView` projections)

### Filter
- `NotificationFilter` -- Supports filtering by `UserID`, `Type`, `Unread` (bool), with `Limit`/`Offset` pagination.

## Application (CQRS)

### Commands
- `CreateCommand` / `CreateHandler` -- Creates a new notification for a user, persists it, and publishes the `NotificationSent` event.
- `DeleteCommand` / `DeleteHandler` -- Hard-deletes a notification by ID.

### Queries
- `GetQuery` / `GetHandler` -- Fetches a single notification view by ID.
- `ListQuery` / `ListHandler` -- Returns a paginated, filterable list of notification views with total count.

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/notifications` | Create a new notification |
| GET | `/notifications` | List notifications (paginated via `limit`/`offset` query params) |
| GET | `/notifications/:id` | Get a single notification by ID |
| DELETE | `/notifications/:id` | Delete a notification |

## Usage
```go
import "gct/internal/notification"
```
