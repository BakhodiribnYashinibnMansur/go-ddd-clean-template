package session

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Revoke revokes a session by its ID.
func (uc *UseCase) Revoke(ctx context.Context, id uuid.UUID) error {
	err := uc.repo.User.SessionRepo.Revoke(ctx, id)
	if err != nil {
		return fmt.Errorf("SessionUseCase - Revoke - uc.repo.User.SessionRepo.Revoke: %w", err)
	}

	return nil
}
