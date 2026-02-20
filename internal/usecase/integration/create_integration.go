package integration

import (
	"context"
	"fmt"

	"gct/internal/domain"

	"github.com/google/uuid"
)

// CreateIntegration business logic.
func (uc *UseCase) CreateIntegration(ctx context.Context, req domain.CreateIntegrationRequest) (*domain.Integration, error) {
	integration := &domain.Integration{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		BaseURL:     req.BaseURL,
		IsActive:    req.IsActive,
		Config:      req.Config,
	}

	err := uc.repo.CreateIntegration(ctx, integration)
	if err != nil {
		return nil, fmt.Errorf("failed to create integration: %w", err)
	}

	return integration, nil
}
