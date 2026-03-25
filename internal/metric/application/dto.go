package application

import (
	"time"

	"github.com/google/uuid"
)

// MetricView is a read-model DTO returned by query handlers.
type MetricView struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	LatencyMs  float64   `json:"latency_ms"`
	IsPanic    bool      `json:"is_panic"`
	PanicError *string   `json:"panic_error,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
