package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/validator"
)

func (uc *UseCase) SignOut(ctx context.Context, in *domain.SignOutIn) error {
	uc.logger.Infoc(ctx, "user sign out started", "input", in)

	// Validate input
	if err := validator.ValidateStruct(in); err != nil {
		return err
	}

	sessionID := in.SessionID

	err := uc.repo.Postgres.User.SessionRepo.Revoke(ctx, &domain.SessionFilter{ID: &sessionID})
	if err != nil {
		uc.logger.Errorc(ctx, "user sign out failed: revoke", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(in)
	}
	uc.logger.Infoc(ctx, "user sign out success")
	return nil
}
