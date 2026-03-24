package client

import (
	"context"

	"gct/internal/shared/domain/consts"
	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/ptrutil"
	"gct/internal/shared/infrastructure/validator"
)

func (uc *UseCase) SignUp(ctx context.Context, in *domain.SignUpIn) (*domain.SignInOut, error) {
	phone := ptrutil.StrVal(in.Phone)
	password := ptrutil.StrVal(in.Password)
	usernameStr := ptrutil.StrVal(in.Username)

	uc.logger.Infoc(ctx, "user sign up started", "input", in)

	// Validate input
	if err := validator.ValidateStruct(in); err != nil {
		return nil, err
	}

	var username *string
	if usernameStr != "" {
		username = &usernameStr
	}

	// Assign default "user" role
	defaultRoleName := consts.RoleUser
	defaultRole, err := uc.repo.Postgres.Authz.Role.Get(ctx, &domain.RoleFilter{Name: &defaultRoleName})
	if err != nil {
		uc.logger.Errorc(ctx, "user sign up failed: get default role", "error", err)
		return nil, apperrors.MapRepoToServiceError(err).WithInput(in)
	}

	user := &domain.User{
		Username:   username,
		Phone:      &phone,
		RoleID:     &defaultRole.ID,
		Active:     true,
		IsApproved: true,
		Attributes: make(map[string]any),
	}

	if err := user.SetPassword(password); err != nil {
		uc.logger.Errorc(ctx, "user sign up failed: set password", "error", err)
		return nil, apperrors.MapRepoToServiceError(err).WithInput(in)
	}

	err = uc.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	uc.logger.Infoc(ctx, "user sign up success, performing automatic sign in")

	signInInput := &domain.SignInIn{
		Login:    &phone,
		Password: &password,
		Session: &domain.SessionIn{
			DeviceID:  in.Session.DeviceID,
			IP:        in.Session.IP,
			UserAgent: in.Session.UserAgent,
		},
	}

	return uc.SignIn(ctx, signInInput)
}
