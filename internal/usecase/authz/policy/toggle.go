package policy

import (
	"context"

	apperrors "gct/pkg/errors"

	"github.com/google/uuid"
)

func (u *UseCase) Toggle(ctx context.Context, id uuid.UUID) error {
	u.logger.Infow("policy toggle started", "id", id)

	err := u.repo.Postgres.Authz.Policy.Toggle(ctx, id)
	if err != nil {
		u.logger.Errorw("policy toggle failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(id)
	}

	u.logger.Infow("policy toggle success")
	return nil
}
