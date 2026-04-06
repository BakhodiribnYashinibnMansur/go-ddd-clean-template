package repository

import (
	"context"

	"gct/internal/context/admin/generic/featureflag/domain/entity"

	"github.com/google/uuid"
)

// FeatureFlagFilter carries optional criteria for querying feature flags.
type FeatureFlagFilter struct {
	Search  *string
	Enabled *bool
	Limit   int64
	Offset  int64
}

// ConditionView is a read-model DTO for a rule group condition.
type ConditionView struct {
	ID        uuid.UUID `json:"id"`
	Attribute string    `json:"attribute"`
	Operator  string    `json:"operator"`
	Value     string    `json:"value"`
}

// RuleGroupView is a read-model DTO for a rule group with its conditions.
type RuleGroupView struct {
	ID         entity.RuleGroupID `json:"id"`
	Name       string             `json:"name"`
	Variation  string             `json:"variation"`
	Priority   int                `json:"priority"`
	Conditions []ConditionView    `json:"conditions"`
	CreatedAt  string             `json:"created_at"`
	UpdatedAt  string             `json:"updated_at"`
}

// FeatureFlagView is a read-model DTO for feature flags.
type FeatureFlagView struct {
	ID                entity.FeatureFlagID `json:"id"`
	Name              string               `json:"name"`
	Key               string               `json:"key"`
	Description       string               `json:"description"`
	FlagType          string               `json:"flag_type"`
	DefaultValue      string               `json:"default_value"`
	RolloutPercentage int                  `json:"rollout_percentage"`
	IsActive          bool                 `json:"is_active"`
	RuleGroups        []RuleGroupView      `json:"rule_groups"`
	CreatedAt         string               `json:"created_at"`
	UpdatedAt         string               `json:"updated_at"`
}

// FeatureFlagReadRepository is the read-side (CQRS query) repository.
type FeatureFlagReadRepository interface {
	FindByID(ctx context.Context, id entity.FeatureFlagID) (*FeatureFlagView, error)
	List(ctx context.Context, filter FeatureFlagFilter) ([]*FeatureFlagView, int64, error)
}

// Evaluator provides feature flag evaluation for application consumers.
type Evaluator interface {
	IsEnabled(ctx context.Context, key string, userAttrs map[string]string) bool
	GetString(ctx context.Context, key string, userAttrs map[string]string) string
	GetInt(ctx context.Context, key string, userAttrs map[string]string) int
	GetFloat(ctx context.Context, key string, userAttrs map[string]string) float64
}
