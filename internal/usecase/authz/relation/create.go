package relation

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Create(ctx context.Context, relation *domain.Relation) error {
	u.logger.WithContext(ctx).Infow("relation create started", "input", relation)

	err := u.repo.Postgres.Authz.Relation.Create(ctx, relation)
	if err != nil {
		u.logger.WithContext(ctx).Errorw("relation create failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(relation)
	}

	u.logger.WithContext(ctx).Infow("relation create success")
	return nil
}
