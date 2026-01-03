package role

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Update(ctx context.Context, role *domain.Role) error {
	u.logger.Infow("role update started", "input", role)

	err := u.repo.Postgres.Authz.Role.Update(ctx, role)
	if err != nil {
		u.logger.Errorw("role update failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(role)
	}
	u.logger.Infow("role update success")
	return nil
}
