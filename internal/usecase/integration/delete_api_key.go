package integration

import (
	"context"

	"github.com/google/uuid"
)

// DeleteAPIKey business logic.
func (uc *UseCase) DeleteAPIKey(ctx context.Context, id uuid.UUID) error {
	return uc.repo.DeleteAPIKey(ctx, id)
}
