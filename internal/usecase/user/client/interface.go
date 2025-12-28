package client

import (
	"context"

	"github.com/evrone/go-clean-template/internal/domain"
)

type (
	CreateInput struct {
		User *domain.User
	}

	UserInput struct {
		ID int64
	}
	UserOutput struct {
		User *domain.User
	}

	UsersInput struct {
		Filter domain.UsersFilter
	}
	UsersOutput struct {
		Users []*domain.User
		Total int
	}

	GetByPhoneInput struct {
		Phone string
	}
	ByPhoneOutput struct {
		User *domain.User
	}

	UpdateInput struct {
		User *domain.User
	}

	DeleteInput struct {
		ID int64
	}

	SignInInput struct {
		Phone     string
		Password  string
		DeviceID  string
		UserAgent string
		IP        string
	}
	SignInOutput struct {
		AccessToken  string
		RefreshToken string
	}

	SignUpInput struct {
		Username string
		Phone    string
		Password string
	}

	SignOutInput struct {
		SessionID string
	}
)

type UseCaseI interface {
	Create(ctx context.Context, in CreateInput) error
	User(ctx context.Context, in UserInput) (UserOutput, error)
	Users(ctx context.Context, in UsersInput) (UsersOutput, error)
	GetByPhone(ctx context.Context, in GetByPhoneInput) (ByPhoneOutput, error)
	Update(ctx context.Context, in UpdateInput) error
	Delete(ctx context.Context, in DeleteInput) error
	SignIn(ctx context.Context, in SignInInput) (SignInOutput, error)
	SignUp(ctx context.Context, in SignUpInput) error
	SignOut(ctx context.Context, in SignOutInput) error
}
