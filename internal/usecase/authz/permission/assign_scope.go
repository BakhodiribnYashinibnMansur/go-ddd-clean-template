package permission

import (
	"context"

	apperrors "gct/pkg/errors"
	"github.com/google/uuid"
)

func (u *UseCase) AssignScope(ctx context.Context, permID uuid.UUID, path, method string) error {
	u.logger.Infow("permission assign scope started", "perm_id", permID, "path", path, "method", method)

	err := u.repo.Postgres.Authz.Permission.AddScope(ctx, permID, path, method)
	if err != nil {
		u.logger.Errorw("permission assign scope failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(map[string]any{"permID": permID, "path": path, "method": method})
	}
	u.logger.Infow("permission assign scope success")
	return nil
}
