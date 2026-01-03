package metric

import (
	"context"

	"go.uber.org/zap"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (uc *UseCase) Gets(ctx context.Context, in *domain.FunctionMetricsFilter) ([]*domain.FunctionMetric, int, error) {
	metrics, count, err := uc.repo.Postgres.Audit.Metric.Gets(ctx, in)
	if err != nil {
		uc.logger.Errorw("failed to retrieve function metrics", zap.Error(err))
		return nil, 0, apperrors.MapRepoToServiceError(ctx, err)
	}
	return metrics, count, nil
}
