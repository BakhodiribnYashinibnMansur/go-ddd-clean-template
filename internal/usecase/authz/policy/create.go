package policy

import (
	"context"

	"gct/consts"
	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Create(ctx context.Context, policy *domain.Policy) error {
	u.logger.WithContext(ctx).Infow("policy create started", "input", policy)

	// Validate conditions keys
	for k := range policy.Conditions {
		if !consts.AllowedPolicyKeys[k] {
			err := apperrors.New(ctx, apperrors.ErrValidation, "invalid policy condition key: "+k).WithInput(map[string]string{"key": k})
			u.logger.WithContext(ctx).Warnw("policy create failed: invalid key", "key", k)
			return err
		}
	}

	err := u.repo.Postgres.Authz.Policy.Create(ctx, policy)
	if err != nil {
		u.logger.WithContext(ctx).Errorw("policy create failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(policy)
	}
	u.logger.WithContext(ctx).Infow("policy create success")
	return nil
}
