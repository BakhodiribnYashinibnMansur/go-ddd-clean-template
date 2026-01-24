package role

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Gets(ctx context.Context, filter *domain.RolesFilter) ([]*domain.Role, int, error) {
	u.logger.WithContext(ctx).Infow("role gets started", "input", filter)

	roles, count, err := u.repo.Postgres.Authz.Role.Gets(ctx, filter)
	if err != nil {
		u.logger.WithContext(ctx).Errorw("role gets failed", "error", err)
		return nil, 0, apperrors.MapRepoToServiceError(err).WithInput(filter)
	}

	u.logger.WithContext(ctx).Infow("role gets success", "count", len(roles), "total", count)
	return roles, count, nil
}
