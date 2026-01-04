package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (uc *UseCase) SignOut(ctx context.Context, in *domain.SignOutIn) error {
	uc.logger.Infow("user sign out started", "input", in)

	sessionID := in.SessionID

	err := uc.repo.Postgres.User.SessionRepo.Revoke(ctx, &domain.SessionFilter{ID: &sessionID})
	if err != nil {
		uc.logger.Errorw("user sign out failed: revoke", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}
	uc.logger.Infow("user sign out success")
	return nil
}
