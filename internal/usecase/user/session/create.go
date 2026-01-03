package session

import (
	"context"
	"time"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/google/uuid"
)

// Create creates a new session.
func (uc *UseCase) Create(ctx context.Context, in *domain.Session) (*domain.Session, error) {
	uc.logger.Infow("session create started", "input", in)

	in.ID = uuid.New()

	// Generate DeviceID if not provided
	if in.DeviceID == uuid.Nil {
		in.DeviceID = uuid.New()
	}

	// Hash refresh token if provided
	if in.RefreshTokenHash != "" {
		// Note: Assuming RefreshTokenHash contains the actual refresh token that needs hashing
		// This might need adjustment based on your requirements
	}

	duration := 24 * time.Hour
	in.ExpiresAt = time.Now().Add(duration)
	in.CreatedAt = time.Now()
	in.UpdatedAt = time.Now()
	in.LastActivity = time.Now()
	in.Revoked = false

	err := uc.repo.Postgres.User.SessionRepo.Create(ctx, in)
	if err != nil {
		uc.logger.Errorw("session create failed", "error", err)
		return nil, apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}

	uc.logger.Infow("session create success", "session_id", in.ID)
	return in, nil
}
