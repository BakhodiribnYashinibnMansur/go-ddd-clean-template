package featureflag

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, f *domain.FeatureFlag) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.FeatureFlag, error)
	List(ctx context.Context, filter domain.FeatureFlagFilter) ([]domain.FeatureFlag, int64, error)
	Update(ctx context.Context, f *domain.FeatureFlag) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UseCaseI interface {
	Create(ctx context.Context, req domain.CreateFeatureFlagRequest) (*domain.FeatureFlag, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.FeatureFlag, error)
	List(ctx context.Context, filter domain.FeatureFlagFilter) ([]domain.FeatureFlag, int64, error)
	Update(ctx context.Context, id uuid.UUID, req domain.UpdateFeatureFlagRequest) (*domain.FeatureFlag, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Toggle(ctx context.Context, id uuid.UUID) (*domain.FeatureFlag, error)
}
