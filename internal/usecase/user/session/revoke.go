package session

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

// Revoke revokes a session.
func (uc *UseCase) Revoke(ctx context.Context, in *domain.SessionFilter) error {
	repo := uc.repo.Postgres.SessionRepo
	err := repo.Revoke(ctx, in)
	if err != nil {
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}
	return nil
}
