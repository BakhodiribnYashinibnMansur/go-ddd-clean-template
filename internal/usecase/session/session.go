package session

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/google/uuid"
)

// UseCase -.
type UseCase struct {
	repo repo.SessionRepo
}

// New -.
func New(r repo.SessionRepo) *UseCase {
	return &UseCase{
		repo: r,
	}
}

// Create creates a new session with the given duration.
func (uc *UseCase) Create(ctx context.Context, s entity.Session, duration time.Duration) (entity.Session, error) {
	s.ExpiresAt = time.Now().Add(duration)
	return uc.repo.Create(ctx, s)
}

// GetByID gets a session by ID and validates it's not expired.
func (uc *UseCase) GetByID(ctx context.Context, id uuid.UUID) (entity.Session, error) {
	s, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return entity.Session{}, fmt.Errorf("SessionUseCase - GetByID - uc.repo.GetByID: %w", err)
	}

	if s.IsExpired() {
		// Delete expired session
		_ = uc.repo.Delete(ctx, id)
		return entity.Session{}, fmt.Errorf("session expired")
	}

	return s, nil
}

// GetByUserID gets all active sessions for a user.
func (uc *UseCase) GetByUserID(ctx context.Context, turonID int64) ([]entity.Session, error) {
	sessions, err := uc.repo.GetByUserID(ctx, turonID)
	if err != nil {
		return nil, fmt.Errorf("SessionUseCase - GetByUserID - uc.repo.GetByUserID: %w", err)
	}

	// Filter out expired sessions
	var activeSessions []entity.Session
	for _, s := range sessions {
		if !s.IsExpired() {
			activeSessions = append(activeSessions, s)
		} else {
			// Optionally delete expired sessions
			_ = uc.repo.Delete(ctx, s.ID)
		}
	}

	return activeSessions, nil
}

// GetOrCreateByDevice gets existing session by device ID or creates a new one.
func (uc *UseCase) GetOrCreateByDevice(ctx context.Context, turonID int64, deviceID uuid.UUID, s entity.Session, duration time.Duration) (entity.Session, error) {
	// Try to get existing session
	existingSession, err := uc.repo.GetByDeviceID(ctx, turonID, deviceID)
	if err == nil {
		// Session exists, check if expired
		if existingSession.IsExpired() {
			// Delete old session and create new one
			_ = uc.repo.Delete(ctx, existingSession.ID)
		} else {
			// Update existing session
			_ = uc.repo.UpdateActivity(ctx, existingSession.ID, s.FCMToken)
			return uc.repo.GetByID(ctx, existingSession.ID)
		}
	}

	// Create new session
	s.ExpiresAt = time.Now().Add(duration)
	return uc.repo.Create(ctx, s)
}

// UpdateActivity updates session's last activity timestamp and FCM token.
func (uc *UseCase) UpdateActivity(ctx context.Context, id uuid.UUID, fcmToken *string) error {
	// Verify session exists and is not expired
	s, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("SessionUseCase - UpdateActivity - uc.repo.GetByID: %w", err)
	}

	if s.IsExpired() {
		_ = uc.repo.Delete(ctx, id)
		return fmt.Errorf("session expired")
	}

	return uc.repo.UpdateActivity(ctx, id, fcmToken)
}

// Delete deletes a session by ID.
func (uc *UseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return uc.repo.Delete(ctx, id)
}

// DeleteByUserID logs out all sessions for a user.
func (uc *UseCase) DeleteByUserID(ctx context.Context, turonID int64) error {
	return uc.repo.DeleteByUserID(ctx, turonID)
}

// CleanupExpired removes all expired sessions.
func (uc *UseCase) CleanupExpired(ctx context.Context) error {
	return uc.repo.DeleteExpired(ctx)
}
