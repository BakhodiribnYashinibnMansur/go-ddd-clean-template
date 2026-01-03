package endpointHistory

import (
	"context"

	"go.uber.org/zap"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (uc *UseCase) Gets(ctx context.Context, in *domain.EndpointHistoriesFilter) ([]*domain.EndpointHistory, int, error) {
	histories, count, err := uc.repo.Postgres.Audit.History.Gets(ctx, in)
	if err != nil {
		uc.logger.Errorw("endpoint history retrieval failed", zap.Error(err))
		return nil, 0, apperrors.MapRepoToServiceError(ctx, err)
	}

	return histories, count, nil
}
