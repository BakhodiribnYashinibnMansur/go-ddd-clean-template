package featureflag

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Create(ctx context.Context, req domain.CreateFeatureFlagRequest) (*domain.FeatureFlag, error) {
	f := &domain.FeatureFlag{
		ID:          uuid.New(),
		Key:         req.Key,
		Name:        req.Name,
		Type:        req.Type,
		Value:       req.Value,
		Description: req.Description,
		IsActive:    req.IsActive,
	}
	if err := uc.repo.Create(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}
