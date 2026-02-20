package integration

import (
	"context"
	"fmt"

	"gct/internal/domain"

	"github.com/google/uuid"
)

// GetIntegration business logic.
func (uc *UseCase) GetIntegration(ctx context.Context, id uuid.UUID) (*domain.IntegrationWithKeys, error) {
	integration, err := uc.repo.GetIntegrationByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get integration: %w", err)
	}

	apiKeys, err := uc.repo.ListAPIKeysByIntegration(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get api keys: %w", err)
	}

	return &domain.IntegrationWithKeys{
		Integration: *integration,
		APIKeys:     apiKeys,
	}, nil
}
