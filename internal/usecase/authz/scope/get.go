package scope

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Get(ctx context.Context, filter *domain.ScopeFilter) (*domain.Scope, error) {
	u.logger.Infoc(ctx, "scope get started", "input", filter)

	scope, err := u.repo.Postgres.Authz.Scope.Get(ctx, filter)
	if err != nil {
		appErr := apperrors.MapRepoToServiceError(err, apperrors.ErrServiceScopeNotFound).WithInput(filter)
		u.logger.Errorc(ctx, "scope get failed", "error", appErr)
		return nil, appErr
	}

	u.logger.Infoc(ctx, "scope get success", "path", scope.Path, "method", scope.Method)
	return scope, nil
}
