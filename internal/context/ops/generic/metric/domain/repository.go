package domain

import (
	"context"
	"time"
)

// MetricFilter carries optional filtering parameters for listing function metrics.
// FromDate/ToDate enable time-range queries for dashboard visualizations.
// Nil pointer fields are treated as "no filter" by the repository implementation.
type MetricFilter struct {
	Name     *string
	IsPanic  *bool
	FromDate *time.Time
	ToDate   *time.Time
	Limit    int64
	Offset   int64
}

// MetricRepository is the write-side repository for the FunctionMetric aggregate.
// List is included on the write side because metric aggregation may need access to full domain objects.
// No FindByID or Delete — metrics are immutable, append-only records.
type MetricRepository interface {
	Save(ctx context.Context, entity *FunctionMetric) error
	List(ctx context.Context, filter MetricFilter) ([]*FunctionMetric, int64, error)
}

// MetricView is a read-model projection for function metrics, optimized for dashboard display.
type MetricView struct {
	ID         MetricID  `json:"id"`
	Name       string    `json:"name"`
	LatencyMs  float64   `json:"latency_ms"`
	IsPanic    bool      `json:"is_panic"`
	PanicError *string   `json:"panic_error,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// MetricReadRepository is the read-side repository returning projected views.
// Only List is provided — individual metric lookup is not a common read path.
type MetricReadRepository interface {
	List(ctx context.Context, filter MetricFilter) ([]*MetricView, int64, error)
}
