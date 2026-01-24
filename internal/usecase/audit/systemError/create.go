package systemError

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
	"go.uber.org/zap"
)

func (uc *UseCase) Create(ctx context.Context, in *domain.SystemError) error {
	err := uc.repo.Postgres.Audit.SystemError.Create(ctx, in)
	if err != nil {
		uc.logger.WithContext(ctx).Errorw("system error creation failed", zap.Error(err))
		return apperrors.MapRepoToServiceError(ctx, err)
	}
	return nil
}
