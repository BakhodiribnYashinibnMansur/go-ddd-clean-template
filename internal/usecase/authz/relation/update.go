package relation

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (u *UseCase) Update(ctx context.Context, relation *domain.Relation) error {
	u.logger.Infow("relation update started", "input", relation)

	err := u.repo.Postgres.Authz.Relation.Update(ctx, relation)
	if err != nil {
		u.logger.Errorw("relation update failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(relation)
	}

	u.logger.Infow("relation update success")
	return nil
}
