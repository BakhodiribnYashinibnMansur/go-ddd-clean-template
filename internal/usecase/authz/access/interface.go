package access

import (
	"context"

	"github.com/google/uuid"

	"gct/internal/domain"
)

type UseCaseI interface {
	Check(ctx context.Context, userID uuid.UUID, session *domain.Session, path, method string, env map[string]any) (bool, error)
}
