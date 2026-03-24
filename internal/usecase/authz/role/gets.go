package role

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (u *UseCase) Gets(ctx context.Context, filter *domain.RolesFilter) ([]*domain.Role, int, error) {
	u.logger.Infoc(ctx, "role gets started", "input", filter)

	roles, count, err := u.repo.Postgres.Authz.Role.Gets(ctx, filter)
	if err != nil {
		u.logger.Errorc(ctx, "role gets failed", "error", err)
		return nil, 0, apperrors.MapRepoToServiceError(err).WithInput(filter)
	}

	u.logger.Infoc(ctx, "role gets success", "count", len(roles), "total", count)
	return roles, count, nil
}
