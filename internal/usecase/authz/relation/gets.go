package relation

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (u *UseCase) Gets(ctx context.Context, filter *domain.RelationsFilter) ([]*domain.Relation, int, error) {
	u.logger.Infow("relation gets started", "input", filter)

	relations, count, err := u.repo.Postgres.Authz.Relation.Gets(ctx, filter)
	if err != nil {
		u.logger.Errorw("relation gets failed", "error", err)
		return nil, 0, apperrors.MapRepoToServiceError(err).WithInput(filter)
	}

	u.logger.Infow("relation gets success", "count", len(relations), "total", count)
	return relations, count, nil
}
