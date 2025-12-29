package session

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (uc *UseCase) Update(ctx context.Context, in *domain.Session) error {
	err := uc.repo.Postgres.SessionRepo.Update(ctx, in)
	if err != nil {
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}
	return nil
}
