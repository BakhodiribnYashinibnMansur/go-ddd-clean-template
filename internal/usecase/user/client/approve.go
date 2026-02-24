package client

import (
	"context"

	apperrors "gct/pkg/errors"
)

// Approve sets is_approved=true for the user with the given ID.
func (uc *UseCase) Approve(ctx context.Context, id string) error {
	uc.logger.Infoc(ctx, "user approve started", "user_id", id)

	if id == "" {
		return apperrors.New(apperrors.ErrInternal, "user id is required")
	}

	err := uc.repo.Postgres.User.Client.Approve(ctx, id)
	if err != nil {
		uc.logger.Errorc(ctx, "user approve failed", "error", err)
		return apperrors.MapRepoToServiceError(err)
	}

	uc.logger.Infoc(ctx, "user approve success", "user_id", id)
	return nil
}
