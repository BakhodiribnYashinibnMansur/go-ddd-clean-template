# Metric

Bounded context for recording function-level performance metrics including latency and panic tracking.

## Domain

### Aggregate Root
- `FunctionMetric` -- Represents a single function execution measurement. Key fields: `name`, `latencyMs`, `isPanic`, `panicError`. Embeds `shared.AggregateRoot`.

### Domain Events
- `FunctionMetricRecorded` -- Raised when a new function metric is recorded via `NewFunctionMetric()`.

### Domain Errors
- `ErrMetricNotFound` -- Returned when a metric cannot be found by ID.

### Repository Interfaces
- `MetricRepository` (write) -- `Save`, `List`
- `MetricReadRepository` (read) -- `List` (returns `MetricView` projections)

### Filter
- `MetricFilter` -- Supports filtering by `Name`, `IsPanic`, `FromDate`, `ToDate`, with `Limit`/`Offset` pagination.

## Application (CQRS)

### Commands
- `RecordMetricCommand` / `RecordMetricHandler` -- Records a new function metric (name, latency, panic status). Publishes `FunctionMetricRecorded` event.

### Queries
- `ListMetricsQuery` / `ListMetricsHandler` -- Lists function metrics with filtering, returns `[]*MetricView` and total count.

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/metrics` | Record a new function metric |
| GET | `/metrics` | List function metrics (paginated via `limit`/`offset` query params) |

## Usage
```go
import "gct/internal/metric"
```
