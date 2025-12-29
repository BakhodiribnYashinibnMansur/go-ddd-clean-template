package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

// Get gets a user.
func (uc *UseCase) Get(ctx context.Context, in *domain.UserFilter) (*domain.User, error) {
	user, err := uc.repo.Postgres.Client.Get(ctx, in)
	if err != nil {
		return nil, apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}
	return user, nil
}
