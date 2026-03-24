package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (uc *UseCase) Gets(ctx context.Context, in *domain.UsersFilter) ([]*domain.User, int, error) {
	uc.logger.Infoc(ctx, "user gets started", "input", in)

	users, total, err := uc.repo.Postgres.User.Client.Gets(ctx, in)
	if err != nil {
		uc.logger.Errorc(ctx, "user gets failed", "error", err)
		return nil, 0, apperrors.MapRepoToServiceError(err).WithInput(in)
	}

	uc.logger.Infoc(ctx, "user gets success", "count", len(users), "total", total)
	return users, total, nil
}
