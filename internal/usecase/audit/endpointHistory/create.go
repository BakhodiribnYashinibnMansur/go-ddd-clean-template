package endpointhistory

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"go.uber.org/zap"
)

func (uc *UseCase) Create(ctx context.Context, in *domain.EndpointHistory) error {
	err := uc.repo.Postgres.Audit.History.Create(ctx, in)
	if err != nil {
		uc.logger.Errorw("endpoint history creation failed", zap.Error(err))
		return apperrors.MapRepoToServiceError(err)
	}
	return nil
}
