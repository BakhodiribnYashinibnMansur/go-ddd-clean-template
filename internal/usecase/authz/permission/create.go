package permission

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (u *UseCase) Create(ctx context.Context, perm *domain.Permission) error {
	u.logger.Infow("permission create started", "input", perm)

	err := u.repo.Postgres.Authz.Permission.Create(ctx, perm)
	if err != nil {
		u.logger.Errorw("permission create failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(perm)
	}

	u.logger.Infow("permission create success")
	return nil
}
