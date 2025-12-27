package session

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/google/uuid"
)

// GetByID gets a session by ID.
func (uc *UseCase) GetByID(ctx context.Context, id uuid.UUID) (domain.Session, error) {
	s, err := uc.repo.User.SessionRepo.GetByID(ctx, id)
	if err != nil {
		return domain.Session{}, fmt.Errorf("SessionUseCase - GetByID - uc.repo.User.SessionRepo.GetByID: %w", err)
	}

	if s.IsExpired() {
		_ = uc.repo.User.SessionRepo.Delete(ctx, id)
		return domain.Session{}, fmt.Errorf("session expired")
	}

	return s, nil
}
