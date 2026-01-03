package permission

import (
	"context"

	"github.com/google/uuid"

	apperrors "gct/pkg/errors"
)

func (u *UseCase) Delete(ctx context.Context, id uuid.UUID) error {
	u.logger.Infow("permission delete started", "id", id)

	err := u.repo.Postgres.Authz.Permission.Delete(ctx, id)
	if err != nil {
		u.logger.Errorw("permission delete failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(id)
	}

	u.logger.Infow("permission delete success")
	return nil
}
