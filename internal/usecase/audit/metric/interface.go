package metric

import (
	"context"

	"gct/internal/domain"
)

type UseCaseI interface {
	Create(ctx context.Context, in *domain.FunctionMetric) error
	Gets(ctx context.Context, in *domain.FunctionMetricsFilter) ([]*domain.FunctionMetric, int, error)
	MeasureSafe(name string) func()
}
