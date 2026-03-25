package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SiteSettingFilter carries filtering parameters for listing site settings.
type SiteSettingFilter struct {
	Key  *string
	Type *string
	Limit  int64
	Offset int64
}

// SiteSettingRepository is the write-side repository for the SiteSetting aggregate.
type SiteSettingRepository interface {
	Save(ctx context.Context, entity *SiteSetting) error
	FindByID(ctx context.Context, id uuid.UUID) (*SiteSetting, error)
	Update(ctx context.Context, entity *SiteSetting) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter SiteSettingFilter) ([]*SiteSetting, int64, error)
}

// SiteSettingView is a read-model DTO for site settings.
type SiteSettingView struct {
	ID          uuid.UUID `json:"id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SiteSettingReadRepository is the read-side repository returning projected views.
type SiteSettingReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*SiteSettingView, error)
	List(ctx context.Context, filter SiteSettingFilter) ([]*SiteSettingView, int64, error)
}
