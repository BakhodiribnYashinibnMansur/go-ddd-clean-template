package integration

import (
	"context"

	"gct/internal/domain"
)

// ListIntegrations business logic.
func (uc *UseCase) ListIntegrations(ctx context.Context, filter domain.IntegrationFilter) ([]domain.Integration, int64, error) {
	return uc.repo.ListIntegrations(ctx, filter)
}
