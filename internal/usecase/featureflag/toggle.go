package featureflag

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Toggle(ctx context.Context, id uuid.UUID) (*domain.FeatureFlag, error) {
	f, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	f.IsActive = !f.IsActive
	if err := uc.repo.Update(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}
