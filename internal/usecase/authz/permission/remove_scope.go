package permission

import (
	"context"

	apperrors "gct/pkg/errors"

	"github.com/google/uuid"
)

func (u *UseCase) RemoveScope(ctx context.Context, permID uuid.UUID, path, method string) error {
	u.logger.WithContext(ctx).Infow("permission remove scope started", "perm_id", permID, "path", path, "method", method)

	err := u.repo.Postgres.Authz.Permission.RemoveScope(ctx, permID, path, method)
	if err != nil {
		u.logger.WithContext(ctx).Errorw("permission remove scope failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(map[string]any{"permID": permID, "path": path, "method": method})
	}
	u.logger.WithContext(ctx).Infow("permission remove scope success")
	return nil
}
