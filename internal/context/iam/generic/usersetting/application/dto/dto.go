package dto

import (
	"time"

	"github.com/google/uuid"
)

// UserSettingView is a read-model DTO returned by query handlers.
type UserSettingView struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
