package role

import (
	"context"

	apperrors "gct/pkg/errors"
	"github.com/google/uuid"
)

func (u *UseCase) Delete(ctx context.Context, id uuid.UUID) error {
	u.logger.WithContext(ctx).Infow("role delete started", "id", id)

	err := u.repo.Postgres.Authz.Role.Delete(ctx, id)
	if err != nil {
		u.logger.WithContext(ctx).Errorw("role delete failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(id)
	}

	u.logger.WithContext(ctx).Infow("role delete success")
	return nil
}
