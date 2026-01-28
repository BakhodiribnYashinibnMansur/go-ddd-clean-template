package policy

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Gets(ctx context.Context, filter *domain.PoliciesFilter) ([]*domain.Policy, int, error) {
	u.logger.Infow("policy gets started", "input", filter)

	policies, count, err := u.repo.Postgres.Authz.Policy.Gets(ctx, filter)
	if err != nil {
		u.logger.Errorw("policy gets failed", "error", err)
		return nil, 0, apperrors.MapRepoToServiceError(err).WithInput(filter)
	}

	u.logger.Infow("policy gets success", "count", len(policies), "total", count)
	return policies, count, nil
}
