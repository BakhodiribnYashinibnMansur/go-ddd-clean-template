package client

import (
	"context"

	apperrors "github.com/evrone/go-clean-template/pkg/errors"
)

func (uc *UseCase) Create(ctx context.Context, in CreateInput) error {
	u := in.User
	// Repo ham pointer qabul qilishi kerak endi
	err := uc.repo.User.Client.Create(ctx, u)
	if err != nil {
		serviceErr := apperrors.AutoSource(
			apperrors.MapRepoToServiceError(ctx, err)).
			WithField("operation", "create_user")

		if u.Username != nil {
			serviceErr.WithField("username", *u.Username)
		}
		if u.Phone != "" {
			serviceErr.WithField("phone", u.Phone)
		}

		return serviceErr
	}
	return nil
}
