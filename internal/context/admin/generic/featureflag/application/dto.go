package application

import (
	"time"

	"gct/internal/context/admin/generic/featureflag/domain"

	"github.com/google/uuid"
)

// ConditionView is a read-model DTO for a rule group condition.
type ConditionView struct {
	ID        uuid.UUID `json:"id"`
	Attribute string    `json:"attribute"`
	Operator  string    `json:"operator"`
	Value     string    `json:"value"`
}

// RuleGroupView is a read-model DTO for a rule group with its conditions.
type RuleGroupView struct {
	ID         domain.RuleGroupID `json:"id"`
	Name       string             `json:"name"`
	Variation  string             `json:"variation"`
	Priority   int                `json:"priority"`
	Conditions []ConditionView    `json:"conditions"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

// FeatureFlagView is a read-model DTO returned by query handlers.
type FeatureFlagView struct {
	ID                domain.FeatureFlagID `json:"id"`
	Name              string               `json:"name"`
	Key               string               `json:"key"`
	Description       string               `json:"description"`
	FlagType          string               `json:"flag_type"`
	DefaultValue      string               `json:"default_value"`
	RolloutPercentage int                  `json:"rollout_percentage"`
	IsActive          bool                 `json:"is_active"`
	RuleGroups        []RuleGroupView      `json:"rule_groups"`
	CreatedAt         time.Time            `json:"created_at"`
	UpdatedAt         time.Time            `json:"updated_at"`
}
