package scope

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Create(ctx context.Context, scope *domain.Scope) error {
	u.logger.Infow("scope create started", "input", scope)

	err := u.repo.Postgres.Authz.Scope.Create(ctx, scope)
	if err != nil {
		u.logger.Errorw("scope create failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(scope)
	}
	u.logger.Infow("scope create success")
	return nil
}
