package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
	"gct/pkg/validator"
)

func (uc *UseCase) SignUp(ctx context.Context, in *domain.SignUpIn) (*domain.SignInOut, error) {
	uc.logger.Infoc(ctx, "user sign up started", "input", in)

	// Validate input
	if err := validator.ValidateStruct(in); err != nil {
		return nil, err
	}

	var username *string
	if in.Username != "" {
		username = &in.Username
	}

	user := &domain.User{
		Username:   username,
		Phone:      &in.Phone,
		Attributes: make(map[string]any),
	}

	if err := user.SetPassword(in.Password); err != nil {
		uc.logger.Errorc(ctx, "user sign up failed: set password", "error", err)
		return nil, apperrors.MapRepoToServiceError(err).WithInput(in)
	}

	err := uc.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	uc.logger.Infoc(ctx, "user sign up success, performing automatic sign in")

	signInInput := &domain.SignInIn{
		Login:    in.Phone,
		Password: in.Password,
	}
	signInInput.Session.DeviceID = in.Session.DeviceID
	signInInput.Session.IP = in.Session.IP
	signInInput.Session.UserAgent = in.Session.UserAgent

	return uc.SignIn(ctx, signInInput)
}
