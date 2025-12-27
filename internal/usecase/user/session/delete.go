package session

import (
	"context"

	"github.com/google/uuid"
)

// Delete terminates a session.
func (uc *UseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return uc.repo.User.SessionRepo.Delete(ctx, id)
}
