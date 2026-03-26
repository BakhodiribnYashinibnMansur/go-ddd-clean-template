package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// FeatureFlag is the aggregate root for feature flag management.
// It combines a boolean kill-switch (enabled) with a rolloutPercentage for gradual rollouts.
// When enabled is false, the flag is completely off regardless of rolloutPercentage.
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

// Toggle flips the enabled state and raises a FlagToggled event.
// This is the preferred method for quick on/off switches; use UpdateDetails for partial field updates.
func (ff *FeatureFlag) Toggle() {
	ff.enabled = !ff.enabled
	ff.Touch()
	ff.AddEvent(NewFlagToggled(ff.ID(), ff.enabled))
}

// UpdateDetails applies partial modifications to the feature flag.
// Nil pointer arguments are treated as "no change." Unlike Toggle, this does not raise a FlagToggled event.
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
