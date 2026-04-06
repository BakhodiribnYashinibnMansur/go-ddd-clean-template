package event

import (
	"time"

	"github.com/google/uuid"
)

// FlagToggled is a domain event emitted when a feature flag's active state changes.
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

// FlagCreated is a domain event emitted when a new feature flag is created.
type FlagCreated struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
}

func NewFlagCreated(id uuid.UUID) FlagCreated {
	return FlagCreated{
		aggregateID: id,
		occurredAt:  time.Now(),
	}
}

func (e FlagCreated) EventName() string      { return "featureflag.created" }
func (e FlagCreated) OccurredAt() time.Time  { return e.occurredAt }
func (e FlagCreated) AggregateID() uuid.UUID { return e.aggregateID }

// FlagUpdated is a domain event emitted when a feature flag is updated.
type FlagUpdated struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
}

func NewFlagUpdated(id uuid.UUID) FlagUpdated {
	return FlagUpdated{
		aggregateID: id,
		occurredAt:  time.Now(),
	}
}

func (e FlagUpdated) EventName() string      { return "featureflag.updated" }
func (e FlagUpdated) OccurredAt() time.Time  { return e.occurredAt }
func (e FlagUpdated) AggregateID() uuid.UUID { return e.aggregateID }

// FlagDeleted is a domain event emitted when a feature flag is deleted.
type FlagDeleted struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
}

func NewFlagDeleted(id uuid.UUID) FlagDeleted {
	return FlagDeleted{
		aggregateID: id,
		occurredAt:  time.Now(),
	}
}

func (e FlagDeleted) EventName() string      { return "featureflag.deleted" }
func (e FlagDeleted) OccurredAt() time.Time  { return e.occurredAt }
func (e FlagDeleted) AggregateID() uuid.UUID { return e.aggregateID }
