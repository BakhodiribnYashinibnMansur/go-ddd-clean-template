package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserSetting struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserSettingFilter struct {
	UserID *uuid.UUID
	Key    *string
}
