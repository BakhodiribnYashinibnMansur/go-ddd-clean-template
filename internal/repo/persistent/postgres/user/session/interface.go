package session

import (
	"context"

	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/google/uuid"
)

// SessionRepoI defines the interface for session repository operations.
type RepoI interface {
	// Create creates a new session.
	Create(ctx context.Context, s domain.Session) error

	// GetByID retrieves a session by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (domain.Session, error)

	// Update updates the session details.
	Update(ctx context.Context, s domain.Session) error

	// Revoke revokes a session by setting its revoked flag to true.
	Revoke(ctx context.Context, id uuid.UUID) error

	// Delete terminates (deletes) a session.
	Delete(ctx context.Context, id uuid.UUID) error
}
