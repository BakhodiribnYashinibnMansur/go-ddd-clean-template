package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (uc *UseCase) Create(ctx context.Context, u *domain.User) error {
	err := uc.repo.Postgres.Client.Create(ctx, u)
	if err != nil {
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(u)
	}
	return nil
}
