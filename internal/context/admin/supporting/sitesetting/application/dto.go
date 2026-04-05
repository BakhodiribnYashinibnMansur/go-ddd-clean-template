package application

import (
	"time"

	"gct/internal/context/admin/supporting/sitesetting/domain"
)

// SiteSettingView is a read-model DTO returned by query handlers.
type SiteSettingView struct {
	ID          domain.SiteSettingID `json:"id"`
	Key         string               `json:"key"`
	Value       string               `json:"value"`
	Type        string               `json:"type"`
	Description string               `json:"description"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}
