package relation

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (u *UseCase) Create(ctx context.Context, relation *domain.Relation) error {
	u.logger.Infow("relation create started", "input", relation)

	err := u.repo.Postgres.Authz.Relation.Create(ctx, relation)
	if err != nil {
		u.logger.Errorw("relation create failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(relation)
	}

	u.logger.Infow("relation create success")
	return nil
}
