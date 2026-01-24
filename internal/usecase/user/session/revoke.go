package session

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

// Revoke revokes a session.
func (uc *UseCase) Revoke(ctx context.Context, in *domain.SessionFilter) error {
	uc.logger.WithContext(ctx).Infow("session revoke started", "input", in)

	repo := uc.repo.Postgres.User.SessionRepo
	err := repo.Revoke(ctx, in)
	if err != nil {
		uc.logger.WithContext(ctx).Errorw("session revoke failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(in)
	}
	uc.logger.WithContext(ctx).Infow("session revoke success")
	return nil
}
