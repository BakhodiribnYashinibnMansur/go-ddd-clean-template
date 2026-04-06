package repository

import (
	"context"
	"time"

	"gct/internal/context/iam/generic/usersetting/domain/entity"

	"github.com/google/uuid"
)

// UserSettingFilter carries optional filtering parameters. Nil fields are ignored by the repository.
type UserSettingFilter struct {
	UserID *uuid.UUID
	Key    *string
	Limit  int64
	Offset int64
}

// UserSettingView is a flat read-model projection for API responses.
type UserSettingView struct {
	ID        entity.UserSettingID `json:"id"`
	UserID    uuid.UUID            `json:"user_id"`
	Key       string               `json:"key"`
	Value     string               `json:"value"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}

// UserSettingReadRepository provides read-only access for listing and detail views.
type UserSettingReadRepository interface {
	FindByID(ctx context.Context, id entity.UserSettingID) (*UserSettingView, error)
	List(ctx context.Context, filter UserSettingFilter) ([]*UserSettingView, int64, error)
}
