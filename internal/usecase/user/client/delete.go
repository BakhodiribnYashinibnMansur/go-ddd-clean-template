package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (uc *UseCase) Delete(ctx context.Context, in *domain.UserFilter) error {
	uc.logger.Infow("user delete started", "input", in)

	if in.ID == nil {
		err := apperrors.New(apperrors.ErrInternal, "user id is required for delete").WithInput(in)
		uc.logger.Errorw("user delete failed: missing id", "error", err)
		return err
	}
	err := uc.repo.Postgres.User.Client.Delete(ctx, *in.ID)
	if err != nil {
		uc.logger.Errorw("user delete failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(in)
	}

	uc.logger.Infow("user delete success")
	return nil
}
