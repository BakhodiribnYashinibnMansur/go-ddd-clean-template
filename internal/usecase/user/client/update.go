package client

import (
	"context"
	"time"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
	"gct/pkg/validation"

	"github.com/google/uuid"
)

func (uc *UseCase) Update(ctx context.Context, u *domain.User) error {
	uc.logger.Infow("user update started", "input", u)

	existing, err := uc.repo.Postgres.User.Client.Get(ctx, &domain.UserFilter{ID: &u.ID})
	if err != nil {
		uc.logger.Errorw("user update failed: get existing", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(u)
	}

	if u.Username != nil {
		existing.Username = u.Username
	}
	if u.Phone != nil && *u.Phone != "" {
		if !validation.IsValidPhone(*u.Phone) {
			return apperrors.NewValidationError("invalid phone format").WithField("phone", *u.Phone)
		}
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
			return apperrors.MapRepoToServiceError(err).WithInput(u)
		}
	}

	err = uc.repo.Postgres.User.Client.Update(ctx, existing)
	if err != nil {
		uc.logger.Errorw("user update failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(u)
	}

	uc.logger.Infow("user update success")
	return nil
}

func (uc *UseCase) SetStatus(ctx context.Context, id uuid.UUID, active bool) error {
	uc.logger.Infow("user set status started", "id", id, "active", active)

	existing, err := uc.repo.Postgres.User.Client.Get(ctx, &domain.UserFilter{ID: &id})
	if err != nil {
		return apperrors.MapRepoToServiceError(err)
	}

	existing.Active = active
	existing.UpdatedAt = time.Now() // Ensure updated_at is refreshed

	err = uc.repo.Postgres.User.Client.Update(ctx, existing)
	if err != nil {
		uc.logger.Errorw("user set status failed", "error", err)
		return apperrors.MapRepoToServiceError(err)
	}

	uc.logger.Infow("user set status success")
	return nil
}
