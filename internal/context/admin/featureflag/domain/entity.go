package domain

import (
	"hash/fnv"
	"sort"
	"time"

	shared "gct/internal/platform/domain"

	"github.com/google/uuid"
)

// FeatureFlag is the aggregate root for feature flag management.
type FeatureFlag struct {
	shared.AggregateRoot
	name              string
	key               string
	description       string
	flagType          string
	defaultValue      string
	rolloutPercentage int
	isActive          bool
	ruleGroups        []*RuleGroup
}

// NewFeatureFlag creates a new FeatureFlag aggregate. isActive defaults to false.
func NewFeatureFlag(name, key, description, flagType, defaultValue string, rolloutPercentage int) *FeatureFlag {
	return &FeatureFlag{
		AggregateRoot:     shared.NewAggregateRoot(),
		name:              name,
		key:               key,
		description:       description,
		flagType:          flagType,
		defaultValue:      defaultValue,
		rolloutPercentage: rolloutPercentage,
		isActive:          false,
	}
}

// ReconstructFeatureFlag rebuilds a FeatureFlag aggregate from persisted data. No events are raised.
func ReconstructFeatureFlag(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
	name, key, description, flagType, defaultValue string,
	rolloutPercentage int,
	isActive bool,
	ruleGroups []*RuleGroup,
) *FeatureFlag {
	return &FeatureFlag{
		AggregateRoot:     shared.NewAggregateRootWithID(id, createdAt, updatedAt, deletedAt),
		name:              name,
		key:               key,
		description:       description,
		flagType:          flagType,
		defaultValue:      defaultValue,
		rolloutPercentage: rolloutPercentage,
		isActive:          isActive,
		ruleGroups:        ruleGroups,
	}
}

// Activate sets the flag to active and raises a FlagToggled event.
func (ff *FeatureFlag) Activate() {
	ff.isActive = true
	ff.Touch()
	ff.AddEvent(NewFlagToggled(ff.ID(), true))
}

// Deactivate sets the flag to inactive and raises a FlagToggled event.
func (ff *FeatureFlag) Deactivate() {
	ff.isActive = false
	ff.Touch()
	ff.AddEvent(NewFlagToggled(ff.ID(), false))
}

// AddRuleGroup appends a rule group to this feature flag.
func (ff *FeatureFlag) AddRuleGroup(rg *RuleGroup) {
	ff.ruleGroups = append(ff.ruleGroups, rg)
}

// UpdateDetails applies partial modifications to the feature flag.
func (ff *FeatureFlag) UpdateDetails(name, key, description *string, flagType *string, defaultValue *string, rolloutPercentage *int, isActive *bool) {
	if name != nil {
		ff.name = *name
	}
	if key != nil {
		ff.key = *key
	}
	if description != nil {
		ff.description = *description
	}
	if flagType != nil {
		ff.flagType = *flagType
	}
	if defaultValue != nil {
		ff.defaultValue = *defaultValue
	}
	if rolloutPercentage != nil {
		ff.rolloutPercentage = *rolloutPercentage
	}
	if isActive != nil {
		ff.isActive = *isActive
	}
	ff.Touch()
}

// Evaluate determines the value of the flag for the given user attributes.
// Inactive flags always return the default value.
// Active flags check rule groups sorted by priority, then rollout percentage.
func (ff *FeatureFlag) Evaluate(userAttrs map[string]string) string {
	if !ff.isActive {
		return ff.defaultValue
	}

	// Sort rule groups by priority (lower number = higher priority)
	sorted := make([]*RuleGroup, len(ff.ruleGroups))
	copy(sorted, ff.ruleGroups)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority() < sorted[j].Priority()
	})

	for _, rg := range sorted {
		if rg.MatchAll(userAttrs) {
			return rg.Variation()
		}
	}

	// Rollout percentage check
	userID, hasUserID := userAttrs["user_id"]
	if hasUserID && ff.isInRollout(userID) {
		return ff.rolloutOnValue()
	}

	return ff.defaultValue
}

// isInRollout uses FNV-1a hash to deterministically bucket users.
func (ff *FeatureFlag) isInRollout(userID string) bool {
	h := fnv.New32a()
	h.Write([]byte(userID + ":" + ff.key))
	bucket := int(h.Sum32() % 100)
	return bucket < ff.rolloutPercentage
}

// rolloutOnValue returns the "on" value for rollout.
func (ff *FeatureFlag) rolloutOnValue() string {
	if ff.flagType == "bool" {
		return "true"
	}
	return ff.defaultValue
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (ff *FeatureFlag) Name() string              { return ff.name }
func (ff *FeatureFlag) Key() string               { return ff.key }
func (ff *FeatureFlag) Description() string       { return ff.description }
func (ff *FeatureFlag) FlagType() string          { return ff.flagType }
func (ff *FeatureFlag) DefaultValue() string      { return ff.defaultValue }
func (ff *FeatureFlag) RolloutPercentage() int    { return ff.rolloutPercentage }
func (ff *FeatureFlag) IsActive() bool            { return ff.isActive }
func (ff *FeatureFlag) RuleGroups() []*RuleGroup  { return ff.ruleGroups }
