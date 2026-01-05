package role

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Update(ctx context.Context, role *domain.Role) error {
	u.logger.WithContext(ctx).Infow("role update started", "input", role)

	err := u.repo.Postgres.Authz.Role.Update(ctx, role)
	if err != nil {
		u.logger.WithContext(ctx).Errorw("role update failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(role)
	}
	u.logger.WithContext(ctx).Infow("role update success")
	return nil
}
