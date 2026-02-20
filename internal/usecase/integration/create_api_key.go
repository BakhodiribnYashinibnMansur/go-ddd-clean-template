package integration

import (
	"context"
	"fmt"

	"gct/consts"
	"gct/internal/domain"

	"github.com/google/uuid"
)

// CreateAPIKey business logic.
func (uc *UseCase) CreateAPIKey(ctx context.Context, req domain.CreateAPIKeyRequest) (*domain.APIKey, string, error) {
	// First check if integration exists
	_, err := uc.repo.GetIntegrationByID(ctx, req.IntegrationID)
	if err != nil {
		return nil, "", fmt.Errorf("integration not found: %w", err)
	}

	apiKey := &domain.APIKey{
		ID:            uuid.New(),
		IntegrationID: req.IntegrationID,
		Name:          req.Name,
		KeyPrefix:     consts.DefaultAPIKeyPrefix,
		IsActive:      true,
		ExpiresAt:     req.ExpiresAt,
	}

	rawKey, err := apiKey.GenerateKey()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate raw key: %w", err)
	}

	err = uc.repo.CreateAPIKey(ctx, apiKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to save api key: %w", err)
	}

	return apiKey, rawKey, nil
}
