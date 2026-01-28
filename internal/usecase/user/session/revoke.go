package session

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

// Revoke revokes a session.
func (uc *UseCase) Revoke(ctx context.Context, in *domain.SessionFilter) error {
	uc.logger.Infoc(ctx, "session revoke started", "input", in)

	err := uc.repo.Postgres.User.SessionRepo.Revoke(ctx, in)
	if err != nil {
		uc.logger.Errorc(ctx, "session revoke failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(in)
	}

	uc.logger.Infoc(ctx, "session revoke success")
	return nil
}
