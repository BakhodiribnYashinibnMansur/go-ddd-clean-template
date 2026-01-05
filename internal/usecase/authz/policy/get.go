package policy

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Get(ctx context.Context, filter *domain.PolicyFilter) (*domain.Policy, error) {
	u.logger.WithContext(ctx).Infow("policy get started", "input", filter)

	policy, err := u.repo.Postgres.Authz.Policy.Get(ctx, filter)
	if err != nil {
		appErr := apperrors.MapRepoToServiceError(ctx, err, apperrors.ErrServicePolicyViolation).WithInput(filter)
		u.logger.WithContext(ctx).Errorw("policy get failed", "error", appErr)
		return nil, appErr
	}

	u.logger.WithContext(ctx).Infow("policy get success", "policy_id", policy.ID)
	return policy, nil
}
