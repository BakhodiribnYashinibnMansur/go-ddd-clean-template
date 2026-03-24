package client

import (
	"context"

	apperrors "gct/internal/shared/infrastructure/errors"
)

// ChangeRole updates the role for the user with the given ID.
func (uc *UseCase) ChangeRole(ctx context.Context, id, role string) error {
	uc.logger.Infoc(ctx, "user change role started", "user_id", id, "role", role)

	if id == "" {
		return apperrors.New(apperrors.ErrInternal, "user id is required")
	}
	if role == "" {
		return apperrors.New(apperrors.ErrInternal, "role is required")
	}

	err := uc.repo.Postgres.User.Client.ChangeRole(ctx, id, role)
	if err != nil {
		uc.logger.Errorc(ctx, "user change role failed", "error", err)
		return apperrors.MapRepoToServiceError(err)
	}

	uc.logger.Infoc(ctx, "user change role success", "user_id", id, "role", role)
	return nil
}
