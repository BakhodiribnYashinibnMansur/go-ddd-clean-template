package role

import (
	"context"

	apperrors "gct/pkg/errors"
	"github.com/google/uuid"
)

func (u *UseCase) AddPermission(ctx context.Context, roleID, permID uuid.UUID) error {
	u.logger.Infow("role add permission started", "role_id", roleID, "perm_id", permID)

	err := u.repo.Postgres.Authz.Role.AddPermission(ctx, roleID, permID)
	if err != nil {
		u.logger.Errorw("role add permission failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(map[string]any{"roleID": roleID, "permID": permID})
	}
	u.logger.Infow("role add permission success")
	return nil
}

func (u *UseCase) RemovePermission(ctx context.Context, roleID, permID uuid.UUID) error {
	u.logger.Infow("role remove permission started", "role_id", roleID, "perm_id", permID)

	err := u.repo.Postgres.Authz.Role.RemovePermission(ctx, roleID, permID)
	if err != nil {
		u.logger.Errorw("role remove permission failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(map[string]any{"roleID": roleID, "permID": permID})
	}

	u.logger.Infow("role remove permission success")
	return nil
}
