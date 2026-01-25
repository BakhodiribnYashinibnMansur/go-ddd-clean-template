package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/google/uuid"
)

// ActivateUser sets the user's Active status to true.
func (uc *UseCase) ActivateUser(ctx context.Context, userID string) error {
	uc.logger.Infoc(ctx, "user activation started", "user_id", userID)

	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	// 1. Get existing user
	existing, err := uc.repo.Postgres.User.Client.Get(ctx, &domain.UserFilter{ID: &uid})
	if err != nil {
		uc.logger.Errorc(ctx, "activation failed: user not found", "error", err)
		return apperrors.MapRepoToServiceError(err)
	}

	// 2. Set IsApproved = true
	existing.IsApproved = true

	// 3. Update in Repo
	err = uc.repo.Postgres.User.Client.Update(ctx, existing)
	if err != nil {
		uc.logger.Errorc(ctx, "activation failed: update error", "error", err)
		return apperrors.MapRepoToServiceError(err)
	}

	uc.logger.Infoc(ctx, "user activation success", "user_id", userID)
	return nil
}
