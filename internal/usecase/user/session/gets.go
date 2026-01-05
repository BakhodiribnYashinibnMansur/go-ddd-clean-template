package session

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (uc *UseCase) Gets(ctx context.Context, in *domain.SessionsFilter) ([]*domain.Session, int, error) {
	uc.logger.WithContext(ctx).Infow("session gets started", "input", in)

	sessions, total, err := uc.repo.Postgres.User.SessionRepo.Gets(ctx, in)
	if err != nil {
		uc.logger.WithContext(ctx).Errorw("session gets failed", "error", err)
		return nil, 0, apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}
	uc.logger.WithContext(ctx).Infow("session gets success", "count", len(sessions), "total", total)
	return sessions, total, nil
}
