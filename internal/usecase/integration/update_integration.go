package integration

import (
	"context"
	"fmt"

	"gct/internal/domain"

	"github.com/google/uuid"
)

// UpdateIntegration business logic.
func (uc *UseCase) UpdateIntegration(ctx context.Context, id uuid.UUID, req domain.UpdateIntegrationRequest) (*domain.Integration, error) {
	integration, err := uc.repo.GetIntegrationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		integration.Name = *req.Name
	}
	if req.Description != nil {
		integration.Description = *req.Description
	}
	if req.BaseURL != nil {
		integration.BaseURL = *req.BaseURL
	}
	if req.IsActive != nil {
		integration.IsActive = *req.IsActive
	}
	if req.Config != nil {
		integration.Config = *req.Config
	}

	err = uc.repo.UpdateIntegration(ctx, integration)
	if err != nil {
		return nil, fmt.Errorf("failed to update integration: %w", err)
	}

	return integration, nil
}
