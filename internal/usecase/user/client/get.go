package client

import (
	"context"

	apperrors "github.com/evrone/go-clean-template/pkg/errors"
)

// User gets a user by ID.
func (uc *UseCase) User(ctx context.Context, in UserInput) (UserOutput, error) {
	user, err := uc.repo.User.Client.User(ctx, in.ID)
	if err != nil {
		return UserOutput{}, apperrors.AutoSource(
			apperrors.MapRepoToServiceError(ctx, err)).
			WithField("operation", "get_user_by_id").
			WithField("user_id", in.ID)
	}
	return UserOutput{User: user}, nil
}
