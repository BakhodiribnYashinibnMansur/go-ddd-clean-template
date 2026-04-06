package dto

import (
	"time"

	"github.com/google/uuid"
)

// SiteSettingView is a read-model DTO returned by query handlers.
type SiteSettingView struct {
	ID          uuid.UUID `json:"id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
