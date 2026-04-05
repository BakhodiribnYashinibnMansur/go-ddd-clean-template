package application

import (
	"time"

	"gct/internal/context/iam/generic/usersetting/domain"

	"github.com/google/uuid"
)

// UserSettingView is a read-model DTO returned by query handlers.
type UserSettingView struct {
	ID        domain.UserSettingID `json:"id"`
	UserID    uuid.UUID            `json:"user_id"`
	Key       string               `json:"key"`
	Value     string               `json:"value"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}
