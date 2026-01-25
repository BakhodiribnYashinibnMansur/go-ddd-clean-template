package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

// GetByPhone gets a user by phone.
func (uc *UseCase) GetByPhone(ctx context.Context, in *domain.UserFilter) (*domain.User, error) {
	uc.logger.Infoc(ctx, "user get by phone started", "input", in)

	if in.Phone == nil || *in.Phone == "" {
		err := apperrors.New(apperrors.ErrServiceInvalidInput, "phone is required").WithInput(in)
		uc.logger.Errorc(ctx, "user get by phone failed: missing phone", "error", err)
		return nil, err
	}

	user, err := uc.repo.Postgres.User.Client.GetByPhone(ctx, *in.Phone)
	if err != nil {
		uc.logger.Errorc(ctx, "user get by phone failed", "error", err)
		return nil, apperrors.MapRepoToServiceError(err).WithInput(in)
	}
	uc.logger.Infoc(ctx, "user get by phone success", "user_id", user.ID)
	return user, nil
}
