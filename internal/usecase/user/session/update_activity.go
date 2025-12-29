package session

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

// UpdateActivity updates session activity using standard Update repo method.
func (uc *UseCase) UpdateActivity(ctx context.Context, in *domain.SessionFilter) error {
	repo := uc.repo.Postgres.SessionRepo
	s, err := repo.GetByID(ctx, in)
	if err != nil {
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}

	if s.IsExpired() || s.Revoked {
		return apperrors.NewServiceError(ctx, apperrors.ErrServiceInvalidInput, "session invalid or revoked").WithInput(in)
	}

	s.UpdateActivity()

	err = repo.Update(ctx, s)
	if err != nil {
		return apperrors.MapRepoToServiceError(ctx, err)
	}
	return nil
}
