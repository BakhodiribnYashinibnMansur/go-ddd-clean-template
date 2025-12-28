package client

import (
	"context"

	apperrors "github.com/evrone/go-clean-template/pkg/errors"
)

func (uc *UseCase) Users(ctx context.Context, in UsersInput) (UsersOutput, error) {
	users, total, err := uc.repo.User.Client.Users(ctx, &in.Filter)
	if err != nil {
		return UsersOutput{}, apperrors.AutoSource(
			apperrors.MapRepoToServiceError(ctx, err)).
			WithField("operation", "get_users")
	}
	return UsersOutput{
		Users: users,
		Total: total,
	}, nil
}
