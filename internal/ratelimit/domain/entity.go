package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// RateLimit is the aggregate root for rate limits.
type RateLimit struct {
	shared.AggregateRoot
	name              string
	rule              string
	requestsPerWindow int
	windowDuration    int
	enabled           bool
}

// NewRateLimit creates a new RateLimit aggregate.
func NewRateLimit(name, rule string, requestsPerWindow, windowDuration int, enabled bool) *RateLimit {
	return &RateLimit{
		AggregateRoot:     shared.NewAggregateRoot(),
		name:              name,
		rule:              rule,
		requestsPerWindow: requestsPerWindow,
		windowDuration:    windowDuration,
		enabled:           enabled,
	}
}

// ReconstructRateLimit rebuilds a RateLimit from persisted data. No events are raised.
func ReconstructRateLimit(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	name, rule string,
	requestsPerWindow, windowDuration int,
	enabled bool,
) *RateLimit {
	return &RateLimit{
		AggregateRoot:     shared.NewAggregateRootWithID(id, createdAt, updatedAt, nil),
		name:              name,
		rule:              rule,
		requestsPerWindow: requestsPerWindow,
		windowDuration:    windowDuration,
		enabled:           enabled,
	}
}

// Update modifies the rate limit fields and raises a RateLimitChanged event.
func (r *RateLimit) Update(name, rule *string, requestsPerWindow, windowDuration *int, enabled *bool) {
	if name != nil {
		r.name = *name
	}
	if rule != nil {
		r.rule = *rule
	}
	if requestsPerWindow != nil {
		r.requestsPerWindow = *requestsPerWindow
	}
	if windowDuration != nil {
		r.windowDuration = *windowDuration
	}
	if enabled != nil {
		r.enabled = *enabled
	}
	r.Touch()
	r.AddEvent(NewRateLimitChanged(r.ID()))
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (r *RateLimit) Name() string              { return r.name }
func (r *RateLimit) Rule() string              { return r.rule }
func (r *RateLimit) RequestsPerWindow() int    { return r.requestsPerWindow }
func (r *RateLimit) WindowDuration() int       { return r.windowDuration }
func (r *RateLimit) Enabled() bool             { return r.enabled }
