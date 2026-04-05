package ports

import (
	"context"

	shared "gct/internal/platform/domain"

	"github.com/google/uuid"
)

// AuthUserLookup is the Anti-Corruption Layer contract used by the authz BC
// (and any future consumer) to obtain minimal authentication data for a user
// without depending on the user BC directly. The user BC provides an adapter
// over its FindUserForAuth query handler at composition time.
type AuthUserLookup interface {
	FindForAuth(ctx context.Context, userID uuid.UUID) (*shared.AuthUser, error)
}
