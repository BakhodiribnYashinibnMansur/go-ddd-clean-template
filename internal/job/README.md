# Job

Bounded context for background and scheduled job execution. Tracks job lifecycle from scheduling through running to completion or failure, with retry support via configurable max attempts.

## Domain

### Aggregate Root
- `Job` -- Represents a background task. Key fields: `taskName`, `status` (PENDING, RUNNING, COMPLETED, FAILED), `payload` (arbitrary JSON input), `result` (arbitrary JSON output), `attempts`, `maxAttempts`, `scheduledAt`, `startedAt`, `completedAt`, `errorMsg`. Embeds `shared.AggregateRoot`.
  - `Start()` -- Transitions status to RUNNING, increments attempts, records start time.
  - `Complete(result)` -- Transitions status to COMPLETED, stores result map, records completion time. Raises `JobCompleted` event.
  - `Fail(errMsg)` -- Transitions status to FAILED with an error message.

### Domain Events
- `JobScheduled` -- Raised when a new job is created via `NewJob()`. Carries aggregate ID and task name.
- `JobCompleted` -- Raised when a job finishes successfully via `Complete()`. Carries aggregate ID and task name.

### Repository Interfaces
- `JobRepository` (write) -- `Save`, `Update`, `FindByID`, `Delete`
- `JobReadRepository` (read) -- `FindByID`, `List` (with `JobFilter`: taskName, status, limit, offset)

## Application (CQRS)

### Commands
- `CreateJobCommand` / `CreateJobHandler` -- Creates a new job with task name, payload, max attempts, and optional scheduled time. Publishes `JobScheduled` event.
- `UpdateJobCommand` / `UpdateJobHandler` -- Transitions a job's status (RUNNING, COMPLETED, FAILED) by calling the appropriate domain method. Publishes domain events on completion.
- `DeleteJobCommand` / `DeleteJobHandler` -- Deletes a job by ID.

### Queries
- `GetJobQuery` / `GetJobHandler` -- Fetches a single job by ID and returns a `JobView` DTO.
- `ListJobsQuery` / `ListJobsHandler` -- Returns a paginated, filtered list of jobs with total count.

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /jobs | Create a new job |
| GET | /jobs | List jobs (paginated via `limit`/`offset` query params) |
| GET | /jobs/:id | Get a single job by ID |
| PATCH | /jobs/:id | Update job status (start, complete, or fail) |
| DELETE | /jobs/:id | Delete a job |

## Usage
```go
import "gct/internal/job"
```
