package access

import (
	"context"

	"gct/internal/domain"
	"github.com/google/uuid"
)

type UseCaseI interface {
	Check(ctx context.Context, userID uuid.UUID, session *domain.Session, path, method string, env map[string]any) (bool, error)
}
