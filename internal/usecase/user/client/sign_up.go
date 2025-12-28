package client

import (
	"context"

	"github.com/evrone/go-clean-template/internal/domain"
)

func (uc *UseCase) SignUp(ctx context.Context, in SignUpInput) error {
	user := &domain.User{
		Username: &in.Username,
		Phone:    in.Phone,
	}

	if err := user.SetPassword(in.Password); err != nil {
		return err
	}

	return uc.Create(ctx, CreateInput{User: user})
}
