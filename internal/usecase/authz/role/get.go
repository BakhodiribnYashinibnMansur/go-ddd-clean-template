package role

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (u *UseCase) Get(ctx context.Context, filter *domain.RoleFilter) (*domain.Role, error) {
	u.logger.Infoc(ctx, "role get started", "input", filter)

	role, err := u.repo.Postgres.Authz.Role.Get(ctx, filter)
	if err != nil {
		appErr := apperrors.MapRepoToServiceError(err, apperrors.ErrServiceRoleNotFound).WithInput(filter)
		u.logger.Errorc(ctx, "role get failed", "error", appErr)
		return nil, appErr
	}

	u.logger.Infoc(ctx, "role get success", "role_id", role.ID)
	return role, nil
}
