package metric

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"go.uber.org/zap"
)

func (uc *UseCase) Gets(ctx context.Context, in *domain.FunctionMetricsFilter) ([]*domain.FunctionMetric, int, error) {
	metrics, count, err := uc.repo.Postgres.Audit.Metric.Gets(ctx, in)
	if err != nil {
		uc.logger.Errorw("failed to retrieve function metrics", zap.Error(err))
		return nil, 0, apperrors.MapRepoToServiceError(err)
	}
	return metrics, count, nil
}
