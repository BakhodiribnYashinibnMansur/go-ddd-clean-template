package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (uc *UseCase) Update(ctx context.Context, u *domain.User) error {
	uc.logger.Infow("user update started", "input", u)

	existing, err := uc.repo.Postgres.User.Client.Get(ctx, &domain.UserFilter{ID: &u.ID})
	if err != nil {
		uc.logger.Errorw("user update failed: get existing", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(u)
	}

	if u.Username != nil {
		existing.Username = u.Username
	}
	if u.Phone != nil && *u.Phone != "" {
		existing.Phone = u.Phone
	}
	if u.Email != nil {
		existing.Email = u.Email
	}
	if u.RoleID != nil {
		existing.RoleID = u.RoleID
	}
	if u.Attributes != nil {
		existing.Attributes = u.Attributes
	}
	if u.Password != "" {
		if err := existing.SetPassword(u.Password); err != nil {
			uc.logger.Errorw("user update failed: set password", "error", err)
			return apperrors.MapRepoToServiceError(ctx, err).WithInput(u)
		}
	}

	err = uc.repo.Postgres.User.Client.Update(ctx, existing)
	if err != nil {
		uc.logger.Errorw("user update failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(u)
	}

	uc.logger.Infow("user update success")
	return nil
}
