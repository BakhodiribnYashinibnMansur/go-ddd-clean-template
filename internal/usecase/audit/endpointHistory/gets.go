package endpointHistory

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
	"go.uber.org/zap"
)

func (uc *UseCase) Gets(ctx context.Context, in *domain.EndpointHistoriesFilter) ([]*domain.EndpointHistory, int, error) {
	histories, count, err := uc.repo.Postgres.Audit.History.Gets(ctx, in)
	if err != nil {
		uc.logger.WithContext(ctx).Errorw("endpoint history retrieval failed", zap.Error(err))
		return nil, 0, apperrors.MapRepoToServiceError(err)
	}

	return histories, count, nil
}
