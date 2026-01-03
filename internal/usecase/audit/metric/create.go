package metric

import (
	"context"

	"go.uber.org/zap"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (uc *UseCase) Create(ctx context.Context, in *domain.FunctionMetric) error {
	err := uc.repo.Postgres.Audit.Metric.Create(ctx, in) // Ensure Postgres repo has Metric field
	if err != nil {
		uc.logger.Errorw("failed to create function metric", zap.Error(err))
		return apperrors.MapRepoToServiceError(ctx, err)
	}
	return nil
}
