package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (uc *UseCase) Delete(ctx context.Context, in *domain.UserFilter) error {
	if in.ID == nil {
		return apperrors.New(ctx, apperrors.ErrInternal, "user id is required for delete").WithInput(in)
	}
	err := uc.repo.Postgres.Client.Delete(ctx, *in.ID)
	if err != nil {
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}
	return nil
}
