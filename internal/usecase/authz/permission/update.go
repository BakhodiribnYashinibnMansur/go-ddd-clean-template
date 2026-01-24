package permission

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Update(ctx context.Context, perm *domain.Permission) error {
	u.logger.WithContext(ctx).Infow("permission update started", "input", perm)

	err := u.repo.Postgres.Authz.Permission.Update(ctx, perm)
	if err != nil {
		u.logger.WithContext(ctx).Errorw("permission update failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(perm)
	}

	u.logger.WithContext(ctx).Infow("permission update success")
	return nil
}
