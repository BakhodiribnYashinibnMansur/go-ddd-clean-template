package role

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Get(ctx context.Context, filter *domain.RoleFilter) (*domain.Role, error) {
	u.logger.WithContext(ctx).Infow("role get started", "input", filter)

	role, err := u.repo.Postgres.Authz.Role.Get(ctx, filter)
	if err != nil {
		appErr := apperrors.MapRepoToServiceError(ctx, err, apperrors.ErrServiceRoleNotFound).WithInput(filter)
		u.logger.WithContext(ctx).Errorw("role get failed", "error", appErr)
		return nil, appErr
	}

	u.logger.WithContext(ctx).Infow("role get success", "role_id", role.ID)
	return role, nil
}
