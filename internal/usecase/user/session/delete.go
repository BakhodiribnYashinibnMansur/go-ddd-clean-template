package session

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

// Delete terminates a session.
func (uc *UseCase) Delete(ctx context.Context, in *domain.SessionFilter) error {
	err := uc.repo.Postgres.SessionRepo.Delete(ctx, in)
	if err != nil {
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}
	return nil
}
