package sitesetting

import (
	"context"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Get(ctx context.Context, filter *domain.SiteSettingFilter) (*domain.SiteSetting, error) {
	query := r.builder.Select(
		schema.SiteSettingID,
		schema.SiteSettingKey,
		schema.SiteSettingValue,
		schema.SiteSettingValueType,
		schema.SiteSettingCategory,
		schema.SiteSettingDescription,
		schema.SiteSettingIsPublic,
		schema.SiteSettingCreatedAt,
		schema.SiteSettingUpdatedAt,
	).From(tableName)

	if filter.ID != nil {
		query = query.Where(squirrel.Eq{schema.SiteSettingID: filter.ID})
	}
	if filter.Key != nil {
		query = query.Where(squirrel.Eq{schema.SiteSettingKey: filter.Key})
	}
	if filter.Category != nil {
		query = query.Where(squirrel.Eq{schema.SiteSettingCategory: filter.Category})
	}
	if filter.IsPublic != nil {
		query = query.Where(squirrel.Eq{schema.SiteSettingIsPublic: filter.IsPublic})
	}

	query = query.Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var setting domain.SiteSetting
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&setting.ID,
		&setting.Key,
		&setting.Value,
		&setting.ValueType,
		&setting.Category,
		&setting.Description,
		&setting.IsPublic,
		&setting.CreatedAt,
		&setting.UpdatedAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return &setting, nil
}
