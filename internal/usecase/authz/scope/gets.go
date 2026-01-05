package scope

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Gets(ctx context.Context, filter *domain.ScopesFilter) ([]*domain.Scope, int, error) {
	u.logger.WithContext(ctx).Infow("scope gets started", "input", filter)

	scopes, count, err := u.repo.Postgres.Authz.Scope.Gets(ctx, filter)
	if err != nil {
		u.logger.WithContext(ctx).Errorw("scope gets failed", "error", err)
		return nil, 0, apperrors.MapRepoToServiceError(ctx, err).WithInput(filter)
	}

	u.logger.WithContext(ctx).Infow("scope gets success", "count", len(scopes), "total", count)
	return scopes, count, nil
}
