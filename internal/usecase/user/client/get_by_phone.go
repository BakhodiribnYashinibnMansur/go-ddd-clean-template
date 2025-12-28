package client

import (
	"context"

	apperrors "github.com/evrone/go-clean-template/pkg/errors"
)

// GetByPhone gets a user by phone.
func (uc *UseCase) GetByPhone(ctx context.Context, in GetByPhoneInput) (ByPhoneOutput, error) {
	user, err := uc.repo.User.Client.GetByPhone(ctx, in.Phone)
	if err != nil {
		return ByPhoneOutput{}, apperrors.AutoSource(
			apperrors.MapRepoToServiceError(ctx, err)).
			WithField("operation", "get_user_by_phone").
			WithField("phone", in.Phone)
	}
	return ByPhoneOutput{User: user}, nil
}
