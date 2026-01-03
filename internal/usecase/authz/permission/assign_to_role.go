package permission

import (
	"context"

	"github.com/google/uuid"

	apperrors "gct/pkg/errors"
)

func (u *UseCase) AssignToRole(ctx context.Context, roleID, permID uuid.UUID) error {
	err := u.repo.Postgres.Authz.Role.AddPermission(ctx, roleID, permID)
	if err != nil {
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(map[string]any{"roleID": roleID, "permID": permID})
	}
	return nil
}
