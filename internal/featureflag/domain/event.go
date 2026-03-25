package domain

import (
	"time"

	"github.com/google/uuid"
)

// FlagToggled is raised when a feature flag's enabled state is toggled.
type FlagToggled struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	Enabled     bool
}

func NewFlagToggled(id uuid.UUID, enabled bool) FlagToggled {
	return FlagToggled{
		aggregateID: id,
		occurredAt:  time.Now(),
		Enabled:     enabled,
	}
}

func (e FlagToggled) EventName() string      { return "featureflag.toggled" }
func (e FlagToggled) OccurredAt() time.Time  { return e.occurredAt }
func (e FlagToggled) AggregateID() uuid.UUID { return e.aggregateID }
