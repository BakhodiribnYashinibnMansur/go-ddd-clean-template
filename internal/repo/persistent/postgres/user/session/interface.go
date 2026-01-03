package session

import (
	"context"

	"gct/internal/domain"
)

// SessionRepoI defines the interface for session repository operations.
type RepoI interface {
	// Create creates a new session.
	Create(ctx context.Context, s *domain.Session) error

	// Get retrieves a session by its ID.
	Get(ctx context.Context, filter *domain.SessionFilter) (*domain.Session, error)

	// Update updates the session details.
	Update(ctx context.Context, s *domain.Session) error

	// Revoke revokes a session by setting its revoked flag to true.
	Revoke(ctx context.Context, filter *domain.SessionFilter) error

	// Delete terminates (deletes) a session.
	Delete(ctx context.Context, filter *domain.SessionFilter) error

	// Gets retrieves sessions based on the provided filter and pagination.
	Gets(ctx context.Context, filter *domain.SessionsFilter) ([]*domain.Session, int, error)
}
