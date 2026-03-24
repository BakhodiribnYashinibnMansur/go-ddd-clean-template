package session

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

// Get gets a session by ID.
func (uc *UseCase) Get(ctx context.Context, in *domain.SessionFilter) (*domain.Session, error) {
	uc.logger.Infoc(ctx, "session get started", "input", in)

	repo := uc.repo.Postgres.User.SessionRepo
	s, err := repo.Get(ctx, in)
	if err != nil {
		uc.logger.Errorc(ctx, "session get failed", "error", err)
		return nil, apperrors.MapRepoToServiceError(err).WithInput(in)
	}

	if s.IsExpired() {
		uc.logger.Warnc(ctx, "session expired, deleting", "session_id", s.ID)
		_ = repo.Delete(ctx, in)
		err := apperrors.NewServiceError(apperrors.ErrServiceInvalidInput, "session expired").WithInput(in)
		uc.logger.Errorc(ctx, "session get failed: expired", "error", err)
		return nil, err
	}

	uc.logger.Infoc(ctx, "session get success", "session_id", s.ID)
	return s, nil
}
