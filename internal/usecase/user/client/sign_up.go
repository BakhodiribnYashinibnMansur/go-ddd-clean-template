package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (uc *UseCase) SignUp(ctx context.Context, in *domain.SignUpIn) error {
	user := &domain.User{
		Username: &in.Username,
		Phone:    in.Phone,
	}

	if err := user.SetPassword(in.Password); err != nil {
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}

	return uc.Create(ctx, user)
}
