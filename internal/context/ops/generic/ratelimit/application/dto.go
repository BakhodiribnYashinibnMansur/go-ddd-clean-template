package application

import (
	"time"

	"gct/internal/context/ops/generic/ratelimit/domain"
)

// RateLimitView is a read-model DTO returned by query handlers.
type RateLimitView struct {
	ID                domain.RateLimitID `json:"id"`
	Name              string             `json:"name"`
	Rule              string             `json:"rule"`
	RequestsPerWindow int                `json:"requests_per_window"`
	WindowDuration    int                `json:"window_duration"`
	Enabled           bool               `json:"enabled"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}
