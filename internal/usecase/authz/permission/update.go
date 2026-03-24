package permission

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (u *UseCase) Update(ctx context.Context, perm *domain.Permission) error {
	u.logger.Infow("permission update started", "input", perm)

	err := u.repo.Postgres.Authz.Permission.Update(ctx, perm)
	if err != nil {
		u.logger.Errorw("permission update failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(perm)
	}

	u.logger.Infow("permission update success")
	return nil
}
