package repository

import (
	"context"
	"time"

	"gct/internal/context/admin/supporting/sitesetting/domain/entity"
)

// SiteSettingView is a read-model projection optimized for API responses. It bypasses
// aggregate reconstruction, returning flat data directly from the database for read performance.
type SiteSettingView struct {
	ID          entity.SiteSettingID `json:"id"`
	Key         string               `json:"key"`
	Value       string               `json:"value"`
	Type        string               `json:"type"`
	Description string               `json:"description"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

// SiteSettingReadRepository is the read-side query interface. It returns lightweight view DTOs
// and should never be used for write operations — use SiteSettingRepository for mutations.
type SiteSettingReadRepository interface {
	FindByID(ctx context.Context, id entity.SiteSettingID) (*SiteSettingView, error)
	List(ctx context.Context, filter SiteSettingFilter) ([]*SiteSettingView, int64, error)
}
