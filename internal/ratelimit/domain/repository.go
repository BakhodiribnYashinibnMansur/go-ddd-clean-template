package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
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
	Save(ctx context.Context, entity *RateLimit) error
	FindByID(ctx context.Context, id uuid.UUID) (*RateLimit, error)
	Update(ctx context.Context, entity *RateLimit) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter RateLimitFilter) ([]*RateLimit, int64, error)
}

// RateLimitView is a read-model projection for rate limits, optimized for admin UI display.
type RateLimitView struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Rule              string    `json:"rule"`
	RequestsPerWindow int       `json:"requests_per_window"`
	WindowDuration    int       `json:"window_duration"`
	Enabled           bool      `json:"enabled"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// RateLimitReadRepository is the read-side repository returning projected views.
// Implementations must return ErrRateLimitNotFound when FindByID yields no result.
type RateLimitReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*RateLimitView, error)
	List(ctx context.Context, filter RateLimitFilter) ([]*RateLimitView, int64, error)
}
