package endpointHistory

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
	"go.uber.org/zap"
)

func (uc *UseCase) Create(ctx context.Context, in *domain.EndpointHistory) error {
	err := uc.repo.Postgres.Audit.History.Create(ctx, in)
	if err != nil {
		uc.logger.WithContext(ctx).Errorw("endpoint history creation failed", zap.Error(err))
		return apperrors.MapRepoToServiceError(ctx, err)
	}
	return nil
}
