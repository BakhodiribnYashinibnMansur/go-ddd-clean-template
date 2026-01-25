package session

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (uc *UseCase) Update(ctx context.Context, in *domain.Session) error {
	uc.logger.Infoc(ctx, "session update started", "input", in)

	err := uc.repo.Postgres.User.SessionRepo.Update(ctx, in)
	if err != nil {
		uc.logger.Errorc(ctx, "session update failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(in)
	}
	uc.logger.Infoc(ctx, "session update success")
	return nil
}
