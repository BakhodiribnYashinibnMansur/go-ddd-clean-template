package session

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/google/uuid"
)

// Create creates a new session.
func (uc *UseCase) Create(ctx context.Context, s *domain.Session) (*domain.Session, error) {
	s.ID = uuid.New()

	duration := 24 * time.Hour
	s.ExpiresAt = time.Now().Add(duration)
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	s.LastActivity = time.Now()
	s.Revoked = false

	err := uc.repo.User.SessionRepo.Create(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("SessionUseCase - Create - uc.repo.User.SessionRepo.Create: %w", err)
	}
	return s, nil
}
