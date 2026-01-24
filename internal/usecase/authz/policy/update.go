package policy

import (
	"context"

	"gct/consts"
	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Update(ctx context.Context, policy *domain.Policy) error {
	u.logger.WithContext(ctx).Infow("policy update started", "input", policy)

	// Validate conditions keys
	if policy.Conditions != nil {
		for k := range policy.Conditions {
			if !consts.AllowedPolicyKeys[k] {
				err := apperrors.New(apperrors.ErrValidation, "invalid policy condition key: "+k).WithInput(map[string]string{"key": k})
				u.logger.WithContext(ctx).Warnw("policy update failed: invalid key", "key", k)
				return err
			}
		}
	}

	err := u.repo.Postgres.Authz.Policy.Update(ctx, policy)
	if err != nil {
		u.logger.WithContext(ctx).Errorw("policy update failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(policy)
	}
	u.logger.WithContext(ctx).Infow("policy update success")
	return nil
}
