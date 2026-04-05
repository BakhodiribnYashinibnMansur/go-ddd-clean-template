package domain

import (
	"context"
	"time"
)

// SiteSettingFilter carries optional filtering parameters for listing site settings.
// Nil pointer fields are ignored by the repository implementation (no filtering on that dimension).
type SiteSettingFilter struct {
	Key    *string
	Type   *string
	Limit  int64
	Offset int64
}

// SiteSettingRepository is the write-side persistence contract for the SiteSetting aggregate.
// Implementations must return ErrSiteSettingNotFound from FindByID when no row matches.
type SiteSettingRepository interface {
	Save(ctx context.Context, entity *SiteSetting) error
	FindByID(ctx context.Context, id SiteSettingID) (*SiteSetting, error)
	Update(ctx context.Context, entity *SiteSetting) error
	Delete(ctx context.Context, id SiteSettingID) error
	List(ctx context.Context, filter SiteSettingFilter) ([]*SiteSetting, int64, error)
}

// SiteSettingView is a read-model projection optimized for API responses. It bypasses
// aggregate reconstruction, returning flat data directly from the database for read performance.
type SiteSettingView struct {
	ID          SiteSettingID `json:"id"`
	Key         string        `json:"key"`
	Value       string        `json:"value"`
	Type        string        `json:"type"`
	Description string        `json:"description"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// SiteSettingReadRepository is the read-side query interface. It returns lightweight view DTOs
// and should never be used for write operations — use SiteSettingRepository for mutations.
type SiteSettingReadRepository interface {
	FindByID(ctx context.Context, id SiteSettingID) (*SiteSettingView, error)
	List(ctx context.Context, filter SiteSettingFilter) ([]*SiteSettingView, int64, error)
}
