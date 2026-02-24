package integration

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

// ToggleIntegration flips the is_active state of an integration.
func (uc *UseCase) ToggleIntegration(ctx context.Context, id uuid.UUID) (*domain.Integration, error) {
	integration, err := uc.repo.GetIntegrationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	integration.IsActive = !integration.IsActive
	if err := uc.repo.UpdateIntegration(ctx, integration); err != nil {
		return nil, err
	}
	return integration, nil
}
