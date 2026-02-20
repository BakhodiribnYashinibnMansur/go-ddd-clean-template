package integration

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

// GetAPIKey business logic.
func (uc *UseCase) GetAPIKey(ctx context.Context, id uuid.UUID) (*domain.APIKey, error) {
	return uc.repo.GetAPIKeyByID(ctx, id)
}
