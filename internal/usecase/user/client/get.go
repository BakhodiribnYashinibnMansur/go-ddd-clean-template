package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

// Get gets a user.
func (uc *UseCase) Get(ctx context.Context, in *domain.UserFilter) (*domain.User, error) {
	uc.logger.Infow("user get started", "input", in)

	user, err := uc.repo.Postgres.User.Client.Get(ctx, in)
	if err != nil {
		uc.logger.Errorw("user get failed", "error", err)
		return nil, apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}

	uc.logger.Infow("user get success", "user_id", user.ID)
	return user, nil
}
