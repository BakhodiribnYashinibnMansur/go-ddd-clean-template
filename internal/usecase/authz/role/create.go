package role

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (u *UseCase) Create(ctx context.Context, role *domain.Role) error {
	u.logger.Infoc(ctx, "role create started", "input", role)

	err := u.repo.Postgres.Authz.Role.Create(ctx, role)
	if err != nil {
		u.logger.Errorc(ctx, "role create failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(role)
	}

	u.logger.Infoc(ctx, "role create success")
	return nil
}
