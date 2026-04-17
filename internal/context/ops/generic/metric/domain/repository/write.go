package repository

import (
	"context"
	"time"

	"gct/internal/context/ops/generic/metric/domain/entity"
	shareddomain "gct/internal/kernel/domain"
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
	Save(ctx context.Context, q shareddomain.Querier, e *entity.FunctionMetric) error
	List(ctx context.Context, filter MetricFilter) ([]*entity.FunctionMetric, int64, error)
}
