package integration

import (
	"context"

	"github.com/google/uuid"
)

// RevokeAPIKey business logic.
func (uc *UseCase) RevokeAPIKey(ctx context.Context, id uuid.UUID) error {
	apiKey, err := uc.repo.GetAPIKeyByID(ctx, id)
	if err != nil {
		return err
	}

	apiKey.IsActive = false
	return uc.repo.UpdateAPIKey(ctx, apiKey)
}
