package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// UserSettingFilter carries filtering parameters for listing user settings.
type UserSettingFilter struct {
	UserID *uuid.UUID
	Key    *string
	Limit  int64
	Offset int64
}

// UserSettingRepository is the write-side repository for the UserSetting aggregate.
type UserSettingRepository interface {
	Upsert(ctx context.Context, entity *UserSetting) error
	FindByUserIDAndKey(ctx context.Context, userID uuid.UUID, key string) (*UserSetting, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// UserSettingView is a read-model DTO for user settings.
type UserSettingView struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserSettingReadRepository is the read-side repository returning projected views.
type UserSettingReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*UserSettingView, error)
	List(ctx context.Context, filter UserSettingFilter) ([]*UserSettingView, int64, error)
}
