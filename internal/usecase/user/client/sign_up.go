package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
	"gct/pkg/validator"
)

func (uc *UseCase) SignUp(ctx context.Context, in *domain.SignUpIn) (*domain.SignInOut, error) {
	// Internal helper to get string from pointer
	strVal := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	phone := strVal(in.Phone)
	password := strVal(in.Password)
	usernameStr := strVal(in.Username)

	uc.logger.Infoc(ctx, "user sign up started", "input", in)

	// Validate input
	if err := validator.ValidateStruct(in); err != nil {
		return nil, err
	}

	var username *string
	if usernameStr != "" {
		username = &usernameStr
	}

	user := &domain.User{
		Username:   username,
		Phone:      &phone,
		Attributes: make(map[string]any),
	}

	if err := user.SetPassword(password); err != nil {
		uc.logger.Errorc(ctx, "user sign up failed: set password", "error", err)
		return nil, apperrors.MapRepoToServiceError(err).WithInput(in)
	}

	err := uc.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	uc.logger.Infoc(ctx, "user sign up success, performing automatic sign in")

	signInInput := &domain.SignInIn{
		Login:    &phone,
		Password: &password,
	}
	signInInput.Session.DeviceID = in.Session.DeviceID
	signInInput.Session.IP = in.Session.IP
	signInInput.Session.UserAgent = in.Session.UserAgent

	return uc.SignIn(ctx, signInInput)
}
