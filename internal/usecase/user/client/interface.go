package client

import (
	"context"

	"github.com/evrone/go-clean-template/internal/domain"
)

type UseCaseI interface {
	Create(ctx context.Context, u domain.User) error
	GetByID(ctx context.Context, id int64) (domain.User, error)
	GetByPhone(ctx context.Context, phone string) (domain.User, error)
	Update(ctx context.Context, u domain.User) error
	Delete(ctx context.Context, id int64) error
}
