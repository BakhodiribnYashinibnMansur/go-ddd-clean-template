package featureflag

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Update(ctx context.Context, id uuid.UUID, req domain.UpdateFeatureFlagRequest) (*domain.FeatureFlag, error) {
	f, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		f.Name = *req.Name
	}
	if req.Type != nil {
		f.Type = *req.Type
	}
	if req.Value != nil {
		f.Value = *req.Value
	}
	if req.Description != nil {
		f.Description = *req.Description
	}
	if req.IsActive != nil {
		f.IsActive = *req.IsActive
	}
	if err := uc.repo.Update(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}
