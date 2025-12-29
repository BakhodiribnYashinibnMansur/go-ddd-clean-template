package session

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (uc *UseCase) Gets(ctx context.Context, in *domain.SessionsFilter) ([]*domain.Session, int, error) {
	sessions, total, err := uc.repo.Postgres.SessionRepo.Gets(ctx, in)
	if err != nil {
		return nil, 0, apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}
	return sessions, total, nil
}
