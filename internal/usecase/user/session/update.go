package session

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (uc *UseCase) Update(ctx context.Context, in *domain.Session) error {
	uc.logger.Infow("session update started", "input", in)

	err := uc.repo.Postgres.User.SessionRepo.Update(ctx, in)
	if err != nil {
		uc.logger.Errorw("session update failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}
	uc.logger.Infow("session update success")
	return nil
}
