package session

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

// UpdateActivity updates session activity using standard Update repo method.
func (uc *UseCase) UpdateActivity(ctx context.Context, in *domain.SessionFilter) error {
	uc.logger.WithContext(ctx).Infow("session update activity started", "input", in)

	repo := uc.repo.Postgres.User.SessionRepo
	s, err := repo.Get(ctx, in)
	if err != nil {
		uc.logger.WithContext(ctx).Errorw("session update activity failed: get", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}

	if s.IsExpired() || s.Revoked {
		err := apperrors.NewServiceError(ctx, apperrors.ErrServiceInvalidInput, "session invalid or revoked").WithInput(in)
		uc.logger.WithContext(ctx).Errorw("session update activity failed: invalid", "error", err)
		return err
	}

	s.UpdateActivity()

	err = repo.Update(ctx, s)
	if err != nil {
		uc.logger.WithContext(ctx).Errorw("session update activity failed: update", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}

	uc.logger.WithContext(ctx).Infow("session update activity success", "session_id", s.ID)
	return nil
}
