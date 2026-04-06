package query

import (
	"context"

	"gct/internal/context/iam/generic/session/application/dto"

	"github.com/google/uuid"
)

// SessionReadRepository defines the read-side persistence contract for sessions.
type SessionReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*dto.SessionView, error)
	List(ctx context.Context, filter dto.SessionsFilter) ([]*dto.SessionView, int64, error)
}
