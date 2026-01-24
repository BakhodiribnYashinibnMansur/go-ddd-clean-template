package access

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

type UseCaseI interface {
	Check(ctx context.Context, userID uuid.UUID, session *domain.Session, path, method string, env map[string]any) (bool, error)
	CheckBatch(ctx context.Context, userID uuid.UUID, session *domain.Session, targets map[string]string, method string, env map[string]any) (map[string]bool, error)
}
