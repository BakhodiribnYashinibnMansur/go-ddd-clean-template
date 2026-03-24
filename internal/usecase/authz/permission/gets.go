package permission

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (u *UseCase) Gets(ctx context.Context, filter *domain.PermissionsFilter) ([]*domain.Permission, int, error) {
	u.logger.Infow("permission gets started", "input", filter)

	perms, count, err := u.repo.Postgres.Authz.Permission.Gets(ctx, filter)
	if err != nil {
		u.logger.Errorw("permission gets failed", "error", err)
		return nil, 0, apperrors.MapRepoToServiceError(err).WithInput(filter)
	}

	u.logger.Infow("permission gets success", "count", len(perms), "total", count)
	return perms, count, nil
}
