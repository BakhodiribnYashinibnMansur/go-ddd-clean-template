package session

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (uc *UseCase) Gets(ctx context.Context, in *domain.SessionsFilter) ([]*domain.Session, int, error) {
	uc.logger.Infoc(ctx, "session gets started", "input", in)

	sessions, total, err := uc.repo.Postgres.User.SessionRepo.Gets(ctx, in)
	if err != nil {
		uc.logger.Errorc(ctx, "session gets failed", "error", err)
		return nil, 0, apperrors.MapRepoToServiceError(err).WithInput(in)
	}
	uc.logger.Infoc(ctx, "session gets success", "count", len(sessions), "total", total)
	return sessions, total, nil
}
