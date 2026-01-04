package policy

import (
	"context"

	apperrors "gct/pkg/errors"
	"github.com/google/uuid"
)

func (u *UseCase) Delete(ctx context.Context, id uuid.UUID) error {
	u.logger.Infow("policy delete started", "id", id)

	err := u.repo.Postgres.Authz.Policy.Delete(ctx, id)
	if err != nil {
		u.logger.Errorw("policy delete failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(id)
	}
	u.logger.Infow("policy delete success")
	return nil
}
