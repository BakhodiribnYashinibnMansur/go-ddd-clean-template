package client

import (
	"context"

	apperrors "github.com/evrone/go-clean-template/pkg/errors"
)

func (uc *UseCase) Update(ctx context.Context, in UpdateInput) error {
	u := in.User
	userData, err := uc.repo.User.Client.User(ctx, u.ID)
	if err != nil {
		return apperrors.AutoSource(
			apperrors.MapRepoToServiceError(ctx, err)).
			WithField("operation", "get_user_for_update").
			WithField("user_id", u.ID)
	}

	// Update fields if provided
	if u.Username != nil {
		userData.Username = u.Username
	}
	if u.Phone != "" {
		userData.Phone = u.Phone
	}
	if u.PasswordHash != "" {
		userData.PasswordHash = u.PasswordHash
	}
	if u.Salt != nil {
		userData.Salt = u.Salt
	}
	if u.LastSeen != nil {
		userData.LastSeen = u.LastSeen
	}

	err = uc.repo.User.Client.Update(ctx, userData)
	if err != nil {
		return apperrors.AutoSource(
			apperrors.MapRepoToServiceError(ctx, err)).
			WithField("operation", "update_user").
			WithField("user_id", u.ID)
	}

	return nil
}
