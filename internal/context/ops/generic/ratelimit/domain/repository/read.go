package repository

import (
	"context"
	"time"

	"gct/internal/context/ops/generic/ratelimit/domain/entity"
)

// RateLimitView is a read-model projection for rate limits, optimized for admin UI display.
type RateLimitView struct {
	ID                entity.RateLimitID `json:"id"`
	Name              string             `json:"name"`
	Rule              string             `json:"rule"`
	RequestsPerWindow int                `json:"requests_per_window"`
	WindowDuration    int                `json:"window_duration"`
	Enabled           bool               `json:"enabled"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

// RateLimitReadRepository is the read-side repository returning projected views.
// Implementations must return ErrRateLimitNotFound when FindByID yields no result.
type RateLimitReadRepository interface {
	FindByID(ctx context.Context, id entity.RateLimitID) (*RateLimitView, error)
	List(ctx context.Context, filter RateLimitFilter) ([]*RateLimitView, int64, error)
}
