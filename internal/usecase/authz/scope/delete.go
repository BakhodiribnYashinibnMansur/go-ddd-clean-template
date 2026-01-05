package scope

import (
	"context"

	apperrors "gct/pkg/errors"
)

func (u *UseCase) Delete(ctx context.Context, path, method string) error {
	u.logger.WithContext(ctx).Infow("scope delete started", "path", path, "method", method)

	err := u.repo.Postgres.Authz.Scope.Delete(ctx, path, method)
	if err != nil {
		u.logger.WithContext(ctx).Errorw("scope delete failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(map[string]string{"path": path, "method": method})
	}
	u.logger.WithContext(ctx).Infow("scope delete success")
	return nil
}
