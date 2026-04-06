package repository

import (
	"context"

	"gct/internal/context/admin/generic/featureflag/domain/entity"
)

// FeatureFlagRepository is the write-side repository for the FeatureFlag aggregate.
type FeatureFlagRepository interface {
	Save(ctx context.Context, entity *entity.FeatureFlag) error
	FindByID(ctx context.Context, id entity.FeatureFlagID) (*entity.FeatureFlag, error)
	FindByKey(ctx context.Context, key string) (*entity.FeatureFlag, error)
	Update(ctx context.Context, entity *entity.FeatureFlag) error
	Delete(ctx context.Context, id entity.FeatureFlagID) error
	FindAll(ctx context.Context) ([]*entity.FeatureFlag, error)
}

// RuleGroupRepository is the write-side repository for RuleGroup entities.
type RuleGroupRepository interface {
	Save(ctx context.Context, rg *entity.RuleGroup) error
	FindByID(ctx context.Context, id entity.RuleGroupID) (*entity.RuleGroup, error)
	Update(ctx context.Context, rg *entity.RuleGroup) error
	Delete(ctx context.Context, id entity.RuleGroupID) error
	FindByFlagID(ctx context.Context, flagID entity.FeatureFlagID) ([]*entity.RuleGroup, error)
	SaveCondition(ctx context.Context, rgID entity.RuleGroupID, c entity.Condition) error
	DeleteConditionsByRuleGroupID(ctx context.Context, rgID entity.RuleGroupID) error
}
