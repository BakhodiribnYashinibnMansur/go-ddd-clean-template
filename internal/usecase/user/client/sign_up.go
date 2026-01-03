package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (uc *UseCase) SignUp(ctx context.Context, in *domain.SignUpIn) error {
	uc.logger.Infow("user sign up started", "input", in)

	user := &domain.User{
		Username:   &in.Username,
		Phone:      &in.Phone,
		Attributes: make(map[string]any),
	}

	if err := user.SetPassword(in.Password); err != nil {
		uc.logger.Errorw("user sign up failed: set password", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}

	err := uc.Create(ctx, user)
	if err != nil {
		return err
	}

	uc.logger.Infow("user sign up success")
	return nil
}
