package scope

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Create(ctx context.Context, scope *domain.Scope) error {
	u.logger.WithContext(ctx).Infow("scope create started", "input", scope)

	err := u.repo.Postgres.Authz.Scope.Create(ctx, scope)
	if err != nil {
		u.logger.WithContext(ctx).Errorw("scope create failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(scope)
	}
	u.logger.WithContext(ctx).Infow("scope create success")
	return nil
}
