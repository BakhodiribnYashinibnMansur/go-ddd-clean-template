# Data Export

Manages user data export requests, tracking them through a lifecycle of statuses (PENDING, PROCESSING, COMPLETED, FAILED). Supports asynchronous export workflows where a request is created, processed in the background, and eventually completed with a downloadable file URL or marked as failed.

## Domain

### Aggregate Root
- `DataExport` -- Represents a single data export request. Key fields: `userID`, `dataType`, `format`, `status`, `fileURL` (optional), `errorMsg` (optional). Embeds `shared.AggregateRoot`.

### Status Constants
- `PENDING` -- Initial state when the export is first requested.
- `PROCESSING` -- The export job is actively running.
- `COMPLETED` -- The export finished successfully; `fileURL` is populated.
- `FAILED` -- The export failed; `errorMsg` is populated.

### Domain Events
- `ExportRequested` -- Raised when a new data export is created. Carries `UserID` and `DataType`.
- `ExportCompleted` -- Raised when an export finishes successfully. Carries `UserID` and `FileURL`.

### Domain Errors
- `ErrDataExportNotFound` -- Returned when a data export cannot be found by ID.

### Repository Interfaces
- `DataExportRepository` (write) -- `Save`, `Update`, `FindByID`, `Delete`
- `DataExportReadRepository` (read) -- `FindByID`, `List` (returns `DataExportView` projections)

### Filter
- `DataExportFilter` -- Supports filtering by `UserID`, `DataType`, `Status`, plus `Limit`/`Offset` pagination.

## Application (CQRS)

### Commands
- `CreateDataExportCommand` / `CreateDataExportHandler` -- Creates a new export request (status starts as PENDING) and publishes the `ExportRequested` event.
- `UpdateDataExportCommand` / `UpdateDataExportHandler` -- Transitions the export through status states (PROCESSING, COMPLETED, FAILED) based on the provided status, file URL, or error message. Publishes domain events on completion.
- `DeleteDataExportCommand` / `DeleteDataExportHandler` -- Deletes a data export record by ID.

### Queries
- `GetDataExportQuery` / `GetDataExportHandler` -- Fetches a single data export view by ID.
- `ListDataExportsQuery` / `ListDataExportsHandler` -- Returns a paginated, filtered list of data export views with total count.

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/data-exports` | Create a new data export request |
| GET | `/data-exports` | List data exports (paginated: `limit`, `offset` query params) |
| GET | `/data-exports/:id` | Get a single data export by ID |
| PATCH | `/data-exports/:id` | Update export status, file URL, or error |
| DELETE | `/data-exports/:id` | Delete a data export |

## Usage
```go
import "gct/internal/dataexport"
```
