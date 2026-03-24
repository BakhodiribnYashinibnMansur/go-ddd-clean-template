package permission

import (
	"context"

	apperrors "gct/internal/shared/infrastructure/errors"
	"github.com/google/uuid"
)

func (u *UseCase) AssignToRole(ctx context.Context, roleID, permID uuid.UUID) error {
	err := u.repo.Postgres.Authz.Role.AddPermission(ctx, roleID, permID)
	if err != nil {
		return apperrors.MapRepoToServiceError(err).WithInput(map[string]any{"roleID": roleID, "permID": permID})
	}
	return nil
}
