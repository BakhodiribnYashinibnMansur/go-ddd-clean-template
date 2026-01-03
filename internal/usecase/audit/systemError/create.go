package systemError

import (
	"context"

	"go.uber.org/zap"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (uc *UseCase) Create(ctx context.Context, in *domain.SystemError) error {
	err := uc.repo.Postgres.Audit.SystemError.Create(ctx, in)
	if err != nil {
		uc.logger.Errorw("system error creation failed", zap.Error(err))
		return apperrors.MapRepoToServiceError(ctx, err)
	}
	return nil
}
