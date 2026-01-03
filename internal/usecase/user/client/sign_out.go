package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/google/uuid"
)

func (uc *UseCase) SignOut(ctx context.Context, in *domain.SignOutIn) error {
	uc.logger.Infow("user sign out started", "input", in)

	sessionID, err := uuid.Parse(in.SessionID)
	if err != nil {
		logErr := apperrors.New(ctx, apperrors.ErrInternal, err.Error()).WithInput(in)
		uc.logger.Errorw("user sign out failed: invalid session id", "error", logErr)
		return logErr
	}

	err = uc.repo.Postgres.User.SessionRepo.Revoke(ctx, &domain.SessionFilter{ID: &sessionID})
	if err != nil {
		uc.logger.Errorw("user sign out failed: revoke", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}
	uc.logger.Infow("user sign out success")
	return nil
}
