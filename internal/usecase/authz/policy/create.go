package policy

import (
	"context"

	"gct/consts"
	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (u *UseCase) Create(ctx context.Context, policy *domain.Policy) error {
	u.logger.Infow("policy create started", "input", policy)

	// Validate conditions keys
	for k := range policy.Conditions {
		if !consts.AllowedPolicyKeys[k] {
			err := apperrors.New(apperrors.ErrValidation, "invalid policy condition key: "+k).WithInput(map[string]string{"key": k})
			u.logger.Warnw("policy create failed: invalid key", "key", k)
			return err
		}
	}

	err := u.repo.Postgres.Authz.Policy.Create(ctx, policy)
	if err != nil {
		u.logger.Errorw("policy create failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(policy)
	}
	u.logger.Infow("policy create success")
	return nil
}
