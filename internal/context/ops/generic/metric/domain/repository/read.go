package repository

import (
	"context"
	"time"

	"gct/internal/context/ops/generic/metric/domain/entity"
)

// MetricView is a read-model projection for function metrics, optimized for dashboard display.
type MetricView struct {
	ID         entity.MetricID `json:"id"`
	Name       string          `json:"name"`
	LatencyMs  float64         `json:"latency_ms"`
	IsPanic    bool            `json:"is_panic"`
	PanicError *string         `json:"panic_error,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

// MetricReadRepository is the read-side repository returning projected views.
// Only List is provided — individual metric lookup is not a common read path.
type MetricReadRepository interface {
	List(ctx context.Context, filter MetricFilter) ([]*MetricView, int64, error)
}
