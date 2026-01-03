package permission

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Get(ctx context.Context, filter *domain.PermissionFilter) (*domain.Permission, error) {
	u.logger.Infow("permission get started", "input", filter)

	perm, err := u.repo.Postgres.Authz.Permission.Get(ctx, filter)
	if err != nil {
		appErr := apperrors.MapRepoToServiceError(ctx, err, apperrors.ErrServicePermissionNotFound).WithInput(filter)
		u.logger.Errorw("permission get failed", "error", appErr)
		return nil, appErr
	}

	u.logger.Infow("permission get success", "perm_id", perm.ID)
	return perm, nil
}
