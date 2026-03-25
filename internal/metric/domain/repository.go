package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MetricFilter carries filtering parameters for listing function metrics.
type MetricFilter struct {
	Name     *string
	IsPanic  *bool
	FromDate *time.Time
	ToDate   *time.Time
	Limit    int64
	Offset   int64
}

// MetricRepository is the write-side repository for the FunctionMetric aggregate.
type MetricRepository interface {
	Save(ctx context.Context, entity *FunctionMetric) error
	List(ctx context.Context, filter MetricFilter) ([]*FunctionMetric, int64, error)
}

// MetricView is a read-model DTO for function metrics.
type MetricView struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	LatencyMs  float64   `json:"latency_ms"`
	IsPanic    bool      `json:"is_panic"`
	PanicError *string   `json:"panic_error,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// MetricReadRepository is the read-side repository returning projected views.
type MetricReadRepository interface {
	List(ctx context.Context, filter MetricFilter) ([]*MetricView, int64, error)
}
