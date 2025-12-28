package session

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/internal/domain"
)

// GetByID gets a session by ID.
func (uc *UseCase) GetByID(ctx context.Context, filter *domain.SessionFilter) (*domain.Session, error) {
	s, err := uc.repo.User.SessionRepo.GetByID(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("SessionUseCase - GetByID - uc.repo.User.SessionRepo.GetByID: %w", err)
	}

	if s.IsExpired() {
		_ = uc.repo.User.SessionRepo.Delete(ctx, filter)
		return nil, fmt.Errorf("session expired")
	}

	return s, nil
}
