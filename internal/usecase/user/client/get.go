package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

// Get gets a user.
func (uc *UseCase) Get(ctx context.Context, in *domain.UserFilter) (*domain.User, error) {
	uc.logger.Infoc(ctx, "user get started", "input", in)

	user, err := uc.repo.Postgres.User.Client.Get(ctx, in)
	if err != nil {
		uc.logger.Errorc(ctx, "user get failed", "error", err)
		return nil, apperrors.MapRepoToServiceError(err).WithInput(in)
	}

	if user == nil {
		uc.logger.Warnc(ctx, "user get failed: user nil", "input", in)
		return nil, apperrors.NewServiceError(apperrors.ErrServiceNotFound, "user not found")
	}

	uc.logger.Infoc(ctx, "user get success", "user_id", user.ID)
	return user, nil
}
