package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

// GetByPhone gets a user by phone.
func (uc *UseCase) GetByPhone(ctx context.Context, in *domain.UserFilter) (*domain.User, error) {
	if in.Phone == nil || *in.Phone == "" {
		return nil, apperrors.New(ctx, apperrors.ErrServiceInvalidInput, "phone is required").WithInput(in)
	}

	user, err := uc.repo.Postgres.Client.GetByPhone(ctx, *in.Phone)
	if err != nil {
		return nil, apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}
	return user, nil
}
