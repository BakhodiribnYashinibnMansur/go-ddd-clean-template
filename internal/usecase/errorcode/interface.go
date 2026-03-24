package errorcode

import (
	"context"
	"gct/internal/domain"
	repo "gct/internal/repo/persistent/postgres/errorcode"
)

type Repository interface {
	Create(ctx context.Context, input repo.CreateErrorCodeInput) (*domain.ErrorCode, error)
	Update(ctx context.Context, code string, input repo.UpdateErrorCodeInput) (*domain.ErrorCode, error)
	GetByCode(ctx context.Context, code string) (*domain.ErrorCode, error)
	List(ctx context.Context) ([]*domain.ErrorCode, error)
	Delete(ctx context.Context, code string) error
}

type UseCaseI interface {
	Create(ctx context.Context, input repo.CreateErrorCodeInput) (*domain.ErrorCode, error)
	Update(ctx context.Context, code string, input repo.UpdateErrorCodeInput) (*domain.ErrorCode, error)
	GetByCode(ctx context.Context, code string) (*domain.ErrorCode, error)
	List(ctx context.Context) ([]*domain.ErrorCode, error)
	Delete(ctx context.Context, code string) error
}
