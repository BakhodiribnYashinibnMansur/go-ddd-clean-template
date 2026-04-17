package repository

import (
	"context"

	"gct/internal/context/admin/supporting/sitesetting/domain/entity"
	shareddomain "gct/internal/kernel/domain"
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
	Save(ctx context.Context, q shareddomain.Querier, entity *entity.SiteSetting) error
	FindByID(ctx context.Context, id entity.SiteSettingID) (*entity.SiteSetting, error)
	Update(ctx context.Context, q shareddomain.Querier, entity *entity.SiteSetting) error
	Delete(ctx context.Context, q shareddomain.Querier, id entity.SiteSettingID) error
	List(ctx context.Context, filter SiteSettingFilter) ([]*entity.SiteSetting, int64, error)
}
