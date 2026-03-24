package role

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (u *UseCase) Update(ctx context.Context, role *domain.Role) error {
	u.logger.Infoc(ctx, "role update started", "input", role)

	err := u.repo.Postgres.Authz.Role.Update(ctx, role)
	if err != nil {
		u.logger.Errorc(ctx, "role update failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(role)
	}
	u.logger.Infoc(ctx, "role update success")
	return nil
}
