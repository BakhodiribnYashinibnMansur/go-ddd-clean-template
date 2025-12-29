package session

import (
	"context"
	"time"

	"github.com/google/uuid"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

// Create creates a new session.
func (uc *UseCase) Create(ctx context.Context, in *domain.Session) (*domain.Session, error) {
	in.ID = uuid.New()

	duration := 24 * time.Hour
	in.ExpiresAt = time.Now().Add(duration)
	in.CreatedAt = time.Now()
	in.UpdatedAt = time.Now()
	in.LastActivity = time.Now()
	in.Revoked = false

	err := uc.repo.Postgres.SessionRepo.Create(ctx, in)
	if err != nil {
		return nil, apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}
	return in, nil
}
