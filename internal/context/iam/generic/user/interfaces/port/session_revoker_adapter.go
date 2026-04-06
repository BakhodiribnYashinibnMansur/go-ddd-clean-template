package port

import (
	"context"

	"gct/internal/context/iam/generic/user/domain"
	"gct/internal/context/iam/generic/user/interfaces/http/middleware"

	"github.com/google/uuid"
)

// SessionRevokerAdapter wraps the domain UserRepository and exposes
// RevokeSessionsByIntegration with uuid.UUID parameters so it satisfies
// the middleware.SessionRevoker interface without a cross-layer coupling.
type SessionRevokerAdapter struct {
	repo domain.UserRepository
}

// NewSessionRevokerAdapter builds the adapter. Construct once in bootstrap.
func NewSessionRevokerAdapter(repo domain.UserRepository) *SessionRevokerAdapter {
	return &SessionRevokerAdapter{repo: repo}
}

// RevokeSessionsByIntegration delegates to the domain repository, converting
// the raw uuid.UUID to domain.UserID.
func (a *SessionRevokerAdapter) RevokeSessionsByIntegration(ctx context.Context, userID uuid.UUID, integrationName string) (int, error) {
	return a.repo.RevokeSessionsByIntegration(ctx, domain.UserID(userID), integrationName)
}

// Compile-time assertion.
var _ middleware.SessionRevoker = (*SessionRevokerAdapter)(nil)
