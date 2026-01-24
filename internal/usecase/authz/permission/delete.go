package permission

import (
	"context"

	apperrors "gct/pkg/errors"
	"github.com/google/uuid"
)

func (u *UseCase) Delete(ctx context.Context, id uuid.UUID) error {
	u.logger.WithContext(ctx).Infow("permission delete started", "id", id)

	err := u.repo.Postgres.Authz.Permission.Delete(ctx, id)
	if err != nil {
		u.logger.WithContext(ctx).Errorw("permission delete failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(id)
	}

	u.logger.WithContext(ctx).Infow("permission delete success")
	return nil
}
