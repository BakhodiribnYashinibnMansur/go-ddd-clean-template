package domain

import (
	"time"

	"github.com/google/uuid"
)

// RateLimitChanged is raised when a rate limit is created or updated.
type RateLimitChanged struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
}

func NewRateLimitChanged(id uuid.UUID) RateLimitChanged {
	return RateLimitChanged{
		aggregateID: id,
		occurredAt:  time.Now(),
	}
}

func (e RateLimitChanged) EventName() string      { return "ratelimit.changed" }
func (e RateLimitChanged) OccurredAt() time.Time   { return e.occurredAt }
func (e RateLimitChanged) AggregateID() uuid.UUID  { return e.aggregateID }
