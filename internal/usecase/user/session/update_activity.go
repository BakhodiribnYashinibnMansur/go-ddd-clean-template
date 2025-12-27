package session

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// UpdateActivity updates session activity using standard Update repo method.
func (uc *UseCase) UpdateActivity(ctx context.Context, id uuid.UUID) error {
	s, err := uc.repo.User.SessionRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("SessionUseCase - UpdateActivity - uc.repo.User.SessionRepo.GetByID: %w", err)
	}

	if s.IsExpired() || s.Revoked {
		return fmt.Errorf("session invalid or revoked")
	}

	s.UpdateActivity()

	err = uc.repo.User.SessionRepo.Update(ctx, s)
	if err != nil {
		return fmt.Errorf("SessionUseCase - UpdateActivity - uc.repo.User.SessionRepo.Update: %w", err)
	}
	return nil
}
