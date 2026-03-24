package scope

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (u *UseCase) Create(ctx context.Context, scope *domain.Scope) error {
	u.logger.Infoc(ctx, "scope create started", "input", scope)

	err := u.repo.Postgres.Authz.Scope.Create(ctx, scope)
	if err != nil {
		u.logger.Errorc(ctx, "scope create failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(scope)
	}
	u.logger.Infoc(ctx, "scope create success")
	return nil
}
