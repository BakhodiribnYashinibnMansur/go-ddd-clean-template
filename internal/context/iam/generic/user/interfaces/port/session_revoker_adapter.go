package port

import (
	"context"

	userentity "gct/internal/context/iam/generic/user/domain/entity"
	userrepo "gct/internal/context/iam/generic/user/domain/repository"
	"gct/internal/context/iam/generic/user/interfaces/http/middleware"

	"github.com/google/uuid"
)

// SessionRevokerAdapter wraps the domain UserRepository and exposes
// RevokeSessionsByIntegration with uuid.UUID parameters so it satisfies
// the middleware.SessionRevoker interface without a cross-layer coupling.
type SessionRevokerAdapter struct {
	repo userrepo.UserRepository
}

// NewSessionRevokerAdapter builds the adapter. Construct once in bootstrap.
func NewSessionRevokerAdapter(repo userrepo.UserRepository) *SessionRevokerAdapter {
	return &SessionRevokerAdapter{repo: repo}
}

// RevokeSessionsByIntegration delegates to the domain repository, converting
// the raw uuid.UUID to userentity.UserID.
func (a *SessionRevokerAdapter) RevokeSessionsByIntegration(ctx context.Context, userID uuid.UUID, integrationName string) (int, error) {
	return a.repo.RevokeSessionsByIntegration(ctx, userentity.UserID(userID), integrationName)
}

// Compile-time assertion.
var _ middleware.SessionRevoker = (*SessionRevokerAdapter)(nil)
