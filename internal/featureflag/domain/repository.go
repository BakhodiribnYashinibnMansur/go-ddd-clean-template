package domain

import (
	"context"

	"github.com/google/uuid"
)

// FeatureFlagFilter carries optional criteria for querying feature flags.
// Search performs a substring match against the flag name. Nil fields mean "no filter."
type FeatureFlagFilter struct {
	Search  *string
	Enabled *bool
	Limit   int64
	Offset  int64
}

// FeatureFlagView is a read-model DTO for feature flags.
type FeatureFlagView struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Enabled           bool      `json:"enabled"`
	RolloutPercentage int       `json:"rollout_percentage"`
	CreatedAt         string    `json:"created_at"`
	UpdatedAt         string    `json:"updated_at"`
}

// FeatureFlagRepository is the write-side repository for the FeatureFlag aggregate.
// Implementations must return ErrFeatureFlagNotFound from FindByID when no row matches.
type FeatureFlagRepository interface {
	Save(ctx context.Context, entity *FeatureFlag) error
	FindByID(ctx context.Context, id uuid.UUID) (*FeatureFlag, error)
	Update(ctx context.Context, entity *FeatureFlag) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// FeatureFlagReadRepository is the read-side (CQRS query) repository.
// It returns pre-projected FeatureFlagView DTOs, bypassing aggregate reconstruction for read performance.
type FeatureFlagReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*FeatureFlagView, error)
	List(ctx context.Context, filter FeatureFlagFilter) ([]*FeatureFlagView, int64, error)
}
