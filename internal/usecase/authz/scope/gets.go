package scope

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (u *UseCase) Gets(ctx context.Context, filter *domain.ScopesFilter) ([]*domain.Scope, int, error) {
	u.logger.Infoc(ctx, "scope gets started", "input", filter)

	scopes, count, err := u.repo.Postgres.Authz.Scope.Gets(ctx, filter)
	if err != nil {
		u.logger.Errorc(ctx, "scope gets failed", "error", err)
		return nil, 0, apperrors.MapRepoToServiceError(err).WithInput(filter)
	}

	u.logger.Infoc(ctx, "scope gets success", "count", len(scopes), "total", count)
	return scopes, count, nil
}
