package role

import (
	"context"

	"github.com/google/uuid"

	apperrors "gct/pkg/errors"
)

func (u *UseCase) Delete(ctx context.Context, id uuid.UUID) error {
	u.logger.Infow("role delete started", "id", id)

	err := u.repo.Postgres.Authz.Role.Delete(ctx, id)
	if err != nil {
		u.logger.Errorw("role delete failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(id)
	}

	u.logger.Infow("role delete success")
	return nil
}
