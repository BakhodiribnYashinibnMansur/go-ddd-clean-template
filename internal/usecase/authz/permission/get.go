package permission

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Get(ctx context.Context, filter *domain.PermissionFilter) (*domain.Permission, error) {
	u.logger.WithContext(ctx).Infow("permission get started", "input", filter)

	perm, err := u.repo.Postgres.Authz.Permission.Get(ctx, filter)
	if err != nil {
		appErr := apperrors.MapRepoToServiceError(err, apperrors.ErrServicePermissionNotFound).WithInput(filter)
		u.logger.WithContext(ctx).Errorw("permission get failed", "error", appErr)
		return nil, appErr
	}

	u.logger.WithContext(ctx).Infow("permission get success", "perm_id", perm.ID)
	return perm, nil
}
