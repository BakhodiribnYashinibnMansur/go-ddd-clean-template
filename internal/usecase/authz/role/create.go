package role

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Create(ctx context.Context, role *domain.Role) error {
	u.logger.Infow("role create started", "input", role)

	err := u.repo.Postgres.Authz.Role.Create(ctx, role)
	if err != nil {
		u.logger.Errorw("role create failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(role)
	}

	u.logger.Infow("role create success")
	return nil
}
