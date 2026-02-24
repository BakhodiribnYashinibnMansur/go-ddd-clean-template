package featureflag

import (
	"context"

	"gct/internal/domain"
)

func (uc *UseCase) List(ctx context.Context, filter domain.FeatureFlagFilter) ([]domain.FeatureFlag, int64, error) {
	return uc.repo.List(ctx, filter)
}
