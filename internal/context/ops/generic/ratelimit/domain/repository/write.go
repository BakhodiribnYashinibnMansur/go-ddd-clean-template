package repository

import (
	"context"

	"gct/internal/context/ops/generic/ratelimit/domain/entity"
)

// RateLimitFilter carries optional filtering parameters for listing rate limits.
// Nil pointer fields are treated as "no filter" by the repository implementation.
type RateLimitFilter struct {
	Name    *string
	Enabled *bool
	Limit   int64
	Offset  int64
}

// RateLimitRepository is the write-side repository for the RateLimit aggregate.
// List is included on the write side because enforcement middleware needs access to full aggregates
// for real-time rate limit evaluation, not just read-model projections.
type RateLimitRepository interface {
	Save(ctx context.Context, entity *entity.RateLimit) error
	FindByID(ctx context.Context, id entity.RateLimitID) (*entity.RateLimit, error)
	Update(ctx context.Context, entity *entity.RateLimit) error
	Delete(ctx context.Context, id entity.RateLimitID) error
	List(ctx context.Context, filter RateLimitFilter) ([]*entity.RateLimit, int64, error)
}
