package relation

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Get(ctx context.Context, filter *domain.RelationFilter) (*domain.Relation, error) {
	u.logger.WithContext(ctx).Infow("relation get started", "input", filter)

	relation, err := u.repo.Postgres.Authz.Relation.Get(ctx, filter)
	if err != nil {
		appErr := apperrors.MapRepoToServiceError(ctx, err, apperrors.ErrServiceRelationNotFound).WithInput(filter)
		u.logger.WithContext(ctx).Errorw("relation get failed", "error", appErr)
		return nil, appErr
	}

	u.logger.WithContext(ctx).Infow("relation get success", "relation_id", relation.ID)
	return relation, nil
}
