package dto

import (
	"time"

	"github.com/google/uuid"
)

// RateLimitView is a read-model DTO returned by query handlers.
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
