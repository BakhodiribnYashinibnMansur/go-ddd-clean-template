package role

import (
	"context"

	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/google/uuid"
)

func (u *UseCase) AddPermission(ctx context.Context, roleID, permID uuid.UUID) error {
	u.logger.Infoc(ctx, "role add permission started", "role_id", roleID, "perm_id", permID)

	err := u.repo.Postgres.Authz.Role.AddPermission(ctx, roleID, permID)
	if err != nil {
		u.logger.Errorc(ctx, "role add permission failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(map[string]any{"roleID": roleID, "permID": permID})
	}
	u.logger.Infoc(ctx, "role add permission success")
	return nil
}

func (u *UseCase) RemovePermission(ctx context.Context, roleID, permID uuid.UUID) error {
	u.logger.Infoc(ctx, "role remove permission started", "role_id", roleID, "perm_id", permID)

	err := u.repo.Postgres.Authz.Role.RemovePermission(ctx, roleID, permID)
	if err != nil {
		u.logger.Errorc(ctx, "role remove permission failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(map[string]any{"roleID": roleID, "permID": permID})
	}

	u.logger.Infoc(ctx, "role remove permission success")
	return nil
}
