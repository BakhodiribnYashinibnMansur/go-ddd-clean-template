package query

import (
	"context"

	appdto "gct/internal/context/iam/generic/session/application"

	"github.com/google/uuid"
)

// SessionReadRepository defines the read-side persistence contract for sessions.
type SessionReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*appdto.SessionView, error)
	List(ctx context.Context, filter appdto.SessionsFilter) ([]*appdto.SessionView, int64, error)
}
