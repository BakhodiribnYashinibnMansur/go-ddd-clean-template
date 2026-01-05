package permission

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Create(ctx context.Context, perm *domain.Permission) error {
	u.logger.WithContext(ctx).Infow("permission create started", "input", perm)

	err := u.repo.Postgres.Authz.Permission.Create(ctx, perm)
	if err != nil {
		u.logger.WithContext(ctx).Errorw("permission create failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(perm)
	}

	u.logger.WithContext(ctx).Infow("permission create success")
	return nil
}
