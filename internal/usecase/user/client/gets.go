package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (uc *UseCase) Gets(ctx context.Context, in *domain.UsersFilter) ([]*domain.User, int, error) {
	uc.logger.Infow("user gets started", "input", in)

	users, total, err := uc.repo.Postgres.User.Client.Gets(ctx, in)
	if err != nil {
		uc.logger.Errorw("user gets failed", "error", err)
		return nil, 0, apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}

	uc.logger.Infow("user gets success", "count", len(users), "total", total)
	return users, total, nil
}
