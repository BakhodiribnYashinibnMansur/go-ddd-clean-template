package integration

import (
	"context"

	"github.com/google/uuid"
)

// DeleteIntegration business logic.
func (uc *UseCase) DeleteIntegration(ctx context.Context, id uuid.UUID) error {
	return uc.repo.DeleteIntegration(ctx, id)
}
