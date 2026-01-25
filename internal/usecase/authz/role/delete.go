package role

import (
	"context"

	apperrors "gct/pkg/errors"

	"github.com/google/uuid"
)

func (u *UseCase) Delete(ctx context.Context, id uuid.UUID) error {
	u.logger.Infoc(ctx, "role delete started", "id", id)

	err := u.repo.Postgres.Authz.Role.Delete(ctx, id)
	if err != nil {
		u.logger.Errorc(ctx, "role delete failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(id)
	}

	u.logger.Infoc(ctx, "role delete success")
	return nil
}
