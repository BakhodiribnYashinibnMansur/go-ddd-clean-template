package ratelimit

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, rl *domain.RateLimit) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.RateLimit, error)
	List(ctx context.Context, filter domain.RateLimitFilter) ([]domain.RateLimit, int64, error)
	Update(ctx context.Context, rl *domain.RateLimit) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UseCaseI interface {
	Create(ctx context.Context, req domain.CreateRateLimitRequest) (*domain.RateLimit, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.RateLimit, error)
	List(ctx context.Context, filter domain.RateLimitFilter) ([]domain.RateLimit, int64, error)
	Update(ctx context.Context, id uuid.UUID, req domain.UpdateRateLimitRequest) (*domain.RateLimit, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Toggle(ctx context.Context, id uuid.UUID) (*domain.RateLimit, error)
}
