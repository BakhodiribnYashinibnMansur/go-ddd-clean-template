package systemError

import (
	"context"

	"go.uber.org/zap"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (uc *UseCase) Gets(ctx context.Context, in *domain.SystemErrorsFilter) ([]*domain.SystemError, int, error) {
	errors, count, err := uc.repo.Postgres.Audit.SystemError.Gets(ctx, in)
	if err != nil {
		uc.logger.Errorw("system error retrieval failed", zap.Error(err))
		return nil, 0, apperrors.MapRepoToServiceError(ctx, err)
	}

	return errors, count, nil
}
