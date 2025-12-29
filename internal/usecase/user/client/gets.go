package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (uc *UseCase) Gets(ctx context.Context, in *domain.UsersFilter) ([]*domain.User, int, error) {
	users, total, err := uc.repo.Postgres.Client.Gets(ctx, in)
	if err != nil {
		return nil, 0, apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}
	return users, total, nil
}
