package integration

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

// ListAPIKeys business logic.
func (uc *UseCase) ListAPIKeys(ctx context.Context, integrationID uuid.UUID) ([]domain.APIKey, error) {
	return uc.repo.ListAPIKeysByIntegration(ctx, integrationID)
}
