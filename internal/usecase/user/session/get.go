package session

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

// Get gets a session by ID.
func (uc *UseCase) Get(ctx context.Context, in *domain.SessionFilter) (*domain.Session, error) {
	repo := uc.repo.Postgres.SessionRepo
	s, err := repo.GetByID(ctx, in)
	if err != nil {
		return nil, apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}

	if s.IsExpired() {
		_ = repo.Delete(ctx, in)
		return nil, apperrors.NewServiceError(ctx, apperrors.ErrServiceInvalidInput, "session expired").WithInput(in)
	}

	return s, nil
}
