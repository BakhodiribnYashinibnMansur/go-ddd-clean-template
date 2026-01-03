package role

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Gets(ctx context.Context, filter *domain.RolesFilter) ([]*domain.Role, int, error) {
	u.logger.Infow("role gets started", "input", filter)

	roles, count, err := u.repo.Postgres.Authz.Role.Gets(ctx, filter)
	if err != nil {
		u.logger.Errorw("role gets failed", "error", err)
		return nil, 0, apperrors.MapRepoToServiceError(ctx, err).WithInput(filter)
	}

	u.logger.Infow("role gets success", "count", len(roles), "total", count)
	return roles, count, nil
}
