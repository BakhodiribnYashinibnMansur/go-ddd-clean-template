package relation

import (
	"context"

	apperrors "gct/pkg/errors"
	"github.com/google/uuid"
)

func (u *UseCase) Delete(ctx context.Context, id uuid.UUID) error {
	u.logger.Infow("relation delete started", "id", id)

	err := u.repo.Postgres.Authz.Relation.Delete(ctx, id)
	if err != nil {
		u.logger.Errorw("relation delete failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(id)
	}

	u.logger.Infow("relation delete success")
	return nil
}
