package repository

import (
	"context"

	"gct/internal/context/iam/generic/usersetting/domain/entity"

	"github.com/google/uuid"
)

// UserSettingRepository is the write-side persistence contract for the UserSetting aggregate.
// Upsert creates or updates based on the (userID, key) natural key — this avoids requiring
// callers to check existence before saving. FindByUserIDAndKey returns ErrUserSettingNotFound on miss.
type UserSettingRepository interface {
	Upsert(ctx context.Context, e *entity.UserSetting) error
	FindByUserIDAndKey(ctx context.Context, userID uuid.UUID, key string) (*entity.UserSetting, error)
	Delete(ctx context.Context, id entity.UserSettingID) error
}
