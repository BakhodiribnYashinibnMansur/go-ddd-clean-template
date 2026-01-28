package sitesetting

import (
	"context"
	"time"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Update(ctx context.Context, setting *domain.SiteSetting) error {
	setting.UpdatedAt = time.Now()

	sql, args, err := r.builder.
		Update(tableName).
		Set(schema.SiteSettingValue, setting.Value).
		Set(schema.SiteSettingValueType, setting.ValueType).
		Set(schema.SiteSettingCategory, setting.Category).
		Set(schema.SiteSettingDescription, setting.Description).
		Set(schema.SiteSettingIsPublic, setting.IsPublic).
		Set(schema.SiteSettingUpdatedAt, setting.UpdatedAt).
		Where(squirrel.Eq{schema.SiteSettingID: setting.ID}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
