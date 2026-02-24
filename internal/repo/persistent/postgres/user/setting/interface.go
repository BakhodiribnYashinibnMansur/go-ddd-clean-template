package setting

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

type RepoI interface {
	Gets(ctx context.Context, userID uuid.UUID) ([]domain.UserSetting, error)
	Upsert(ctx context.Context, s *domain.UserSetting) error
	Delete(ctx context.Context, userID uuid.UUID, key string) error
}
