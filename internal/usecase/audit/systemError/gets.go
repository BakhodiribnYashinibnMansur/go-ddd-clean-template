package systemerror

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"go.uber.org/zap"
)

func (uc *UseCase) Gets(ctx context.Context, in *domain.SystemErrorsFilter) ([]*domain.SystemError, int, error) {
	errors, count, err := uc.repo.Postgres.Audit.SystemError.Gets(ctx, in)
	if err != nil {
		uc.logger.Errorw("system error retrieval failed", zap.Error(err))
		return nil, 0, apperrors.MapRepoToServiceError(err)
	}

	return errors, count, nil
}
