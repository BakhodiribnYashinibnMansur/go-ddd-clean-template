package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
	"gct/pkg/validator"
)

func (uc *UseCase) SignUp(ctx context.Context, in *domain.SignUpIn) (*domain.SignInOut, error) {
	uc.logger.WithContext(ctx).Infow("user sign up started", "input", in)

	// Validate input
	if err := validator.ValidateStruct(ctx, in); err != nil {
		return nil, err
	}

	user := &domain.User{
		Username:   &in.Username,
		Phone:      &in.Phone,
		Attributes: make(map[string]any),
	}

	if err := user.SetPassword(in.Password); err != nil {
		uc.logger.WithContext(ctx).Errorw("user sign up failed: set password", "error", err)
		return nil, apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}

	err := uc.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	uc.logger.WithContext(ctx).Infow("user sign up success, performing automatic sign in")

	return uc.SignIn(ctx, &domain.SignInIn{
		Phone:     in.Phone,
		Password:  in.Password,
		DeviceID:  in.DeviceID,
		IP:        in.IP,
		UserAgent: in.UserAgent,
	})
}
