package client

import (
	"context"

	"gct/internal/domain"
)

type UseCaseI interface {
	Create(ctx context.Context, in *domain.User) error
	Get(ctx context.Context, in *domain.UserFilter) (*domain.User, error)
	Gets(ctx context.Context, in *domain.UsersFilter) ([]*domain.User, int, error)
	Update(ctx context.Context, in *domain.User) error
	Delete(ctx context.Context, in *domain.UserFilter) error
	SignIn(ctx context.Context, in *domain.SignInIn) (*domain.SignInOut, error)
	SignUp(ctx context.Context, in *domain.SignUpIn) (*domain.SignInOut, error)
	SignOut(ctx context.Context, in *domain.SignOutIn) error
	RotateSession(ctx context.Context, in *domain.RefreshIn) (*domain.SignInOut, error)
	GetByPhone(ctx context.Context, in *domain.UserFilter) (*domain.User, error)
}
