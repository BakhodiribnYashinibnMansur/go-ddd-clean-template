package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// FeatureFlag is the aggregate root for feature flag management.
type FeatureFlag struct {
	shared.AggregateRoot
	name              string
	description       string
	enabled           bool
	rolloutPercentage int
}

// NewFeatureFlag creates a new FeatureFlag aggregate.
func NewFeatureFlag(name, description string, enabled bool, rolloutPercentage int) *FeatureFlag {
	ff := &FeatureFlag{
		AggregateRoot:     shared.NewAggregateRoot(),
		name:              name,
		description:       description,
		enabled:           enabled,
		rolloutPercentage: rolloutPercentage,
	}
	return ff
}

// ReconstructFeatureFlag rebuilds a FeatureFlag aggregate from persisted data. No events are raised.
func ReconstructFeatureFlag(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
	name, description string,
	enabled bool,
	rolloutPercentage int,
) *FeatureFlag {
	return &FeatureFlag{
		AggregateRoot:     shared.NewAggregateRootWithID(id, createdAt, updatedAt, deletedAt),
		name:              name,
		description:       description,
		enabled:           enabled,
		rolloutPercentage: rolloutPercentage,
	}
}

// Toggle flips the Enabled flag and raises a FlagToggled event.
func (ff *FeatureFlag) Toggle() {
	ff.enabled = !ff.enabled
	ff.Touch()
	ff.AddEvent(NewFlagToggled(ff.ID(), ff.enabled))
}

// UpdateDetails updates mutable fields.
func (ff *FeatureFlag) UpdateDetails(name, description *string, enabled *bool, rolloutPercentage *int) {
	if name != nil {
		ff.name = *name
	}
	if description != nil {
		ff.description = *description
	}
	if enabled != nil {
		ff.enabled = *enabled
	}
	if rolloutPercentage != nil {
		ff.rolloutPercentage = *rolloutPercentage
	}
	ff.Touch()
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (ff *FeatureFlag) Name() string              { return ff.name }
func (ff *FeatureFlag) Description() string       { return ff.description }
func (ff *FeatureFlag) Enabled() bool             { return ff.enabled }
func (ff *FeatureFlag) RolloutPercentage() int    { return ff.rolloutPercentage }
