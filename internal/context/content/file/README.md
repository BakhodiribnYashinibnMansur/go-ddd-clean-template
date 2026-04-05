# File

Bounded context for tracking uploaded file metadata. Files are intentionally immutable after creation -- they can be uploaded and queried but never updated or deleted through this context. The `uploadedBy` field is nullable to support anonymous or system-generated uploads.

## Domain

### Aggregate Root
- `File` -- Immutable file metadata record with fields: `name` (storage name), `originalName` (user-facing name), `mimeType`, `size` (bytes), `path` (storage path), `url` (public URL), and optional `uploadedBy` (user UUID).

### Domain Events
- `FileUploaded` -- Raised when a new file record is created. Carries the aggregate ID, name, MIME type, and size.

### Domain Errors
- `ErrFileNotFound` -- Returned when a requested file does not exist.

### Repository Interfaces
- `FileRepository` (write) -- `Save` (write-only; no update or delete)
- `FileReadRepository` (read) -- `FindByID`, `List` (returns `FileView` projections)

### Filter
- `FileFilter` -- Supports filtering by `Name` and `MimeType`, with `Limit`/`Offset` pagination.

## Application (CQRS)

### Commands
- `CreateFileCommand` / `CreateFileHandler` -- Creates a new file metadata record, persists it, and publishes the `FileUploaded` event.

### Queries
- `GetFileQuery` / `GetFileHandler` -- Fetches a single file view by ID.
- `ListFilesQuery` / `ListFilesHandler` -- Returns a paginated, filterable list of file views with total count.

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/files` | Create a new file metadata record |
| GET | `/files` | List files (paginated via `limit`/`offset` query params) |
| GET | `/files/:id` | Get a single file by ID |

## Usage
```go
import "gct/internal/file"
```
