package role

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Create(ctx context.Context, role *domain.Role) error {
	u.logger.WithContext(ctx).Infow("role create started", "input", role)

	err := u.repo.Postgres.Authz.Role.Create(ctx, role)
	if err != nil {
		u.logger.WithContext(ctx).Errorw("role create failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(role)
	}

	u.logger.WithContext(ctx).Infow("role create success")
	return nil
}
