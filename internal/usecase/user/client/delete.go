package client

import (
	"context"

	apperrors "github.com/evrone/go-clean-template/pkg/errors"
)

func (uc *UseCase) Delete(ctx context.Context, in DeleteInput) error {
	err := uc.repo.User.Client.Delete(ctx, in.ID)
	if err != nil {
		return apperrors.AutoSource(
			apperrors.MapRepoToServiceError(ctx, err)).
			WithField("operation", "delete_user").
			WithField("user_id", in.ID)
	}
	return nil
}
