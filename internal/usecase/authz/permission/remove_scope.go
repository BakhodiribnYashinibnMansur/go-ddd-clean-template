package permission

import (
	"context"

	"github.com/google/uuid"

	apperrors "gct/pkg/errors"
)

func (u *UseCase) RemoveScope(ctx context.Context, permID uuid.UUID, path, method string) error {
	u.logger.Infow("permission remove scope started", "perm_id", permID, "path", path, "method", method)

	err := u.repo.Postgres.Authz.Permission.RemoveScope(ctx, permID, path, method)
	if err != nil {
		u.logger.Errorw("permission remove scope failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(map[string]any{"permID": permID, "path": path, "method": method})
	}
	u.logger.Infow("permission remove scope success")
	return nil
}
