package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// UserSettingFilter carries optional filtering parameters. Nil fields are ignored by the repository.
type UserSettingFilter struct {
	UserID *uuid.UUID
	Key    *string
	Limit  int64
	Offset int64
}

// UserSettingRepository is the write-side persistence contract for the UserSetting aggregate.
// Upsert creates or updates based on the (userID, key) natural key — this avoids requiring
// callers to check existence before saving. FindByUserIDAndKey returns ErrUserSettingNotFound on miss.
type UserSettingRepository interface {
	Upsert(ctx context.Context, entity *UserSetting) error
	FindByUserIDAndKey(ctx context.Context, userID uuid.UUID, key string) (*UserSetting, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// UserSettingView is a flat read-model projection for API responses.
type UserSettingView struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserSettingReadRepository provides read-only access for listing and detail views.
type UserSettingReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*UserSettingView, error)
	List(ctx context.Context, filter UserSettingFilter) ([]*UserSettingView, int64, error)
}
