package session

import (
	"context"

	"github.com/evrone/go-clean-template/internal/domain"
)

// SessionRepoI defines the interface for session repository operations.
type RepoI interface {
	// Create creates a new session.
	Create(ctx context.Context, s *domain.Session) error

	// GetByID retrieves a session by its ID.
	GetByID(ctx context.Context, filter *domain.SessionFilter) (*domain.Session, error)

	// Update updates the session details.
	Update(ctx context.Context, s *domain.Session) error

	// Revoke revokes a session by setting its revoked flag to true.
	Revoke(ctx context.Context, filter *domain.SessionFilter) error

	// Delete terminates (deletes) a session.
	Delete(ctx context.Context, filter *domain.SessionFilter) error
}
