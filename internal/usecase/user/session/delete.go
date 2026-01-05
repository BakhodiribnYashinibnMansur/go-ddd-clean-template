package session

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

// Delete terminates a session.
func (uc *UseCase) Delete(ctx context.Context, in *domain.SessionFilter) error {
	uc.logger.WithContext(ctx).Infow("session delete started", "input", in)

	err := uc.repo.Postgres.User.SessionRepo.Delete(ctx, in)
	if err != nil {
		uc.logger.WithContext(ctx).Errorw("session delete failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}
	uc.logger.WithContext(ctx).Infow("session delete success")
	return nil
}
