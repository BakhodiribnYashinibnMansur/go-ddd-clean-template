package sitesetting

import (
	"context"
	"time"

	"gct/consts"
	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Update(ctx context.Context, setting *domain.SiteSetting) error {
	setting.UpdatedAt = time.Now()

	sql, args, err := r.builder.
		Update(tableName).
		Set("value", setting.Value).
		Set("value_type", setting.ValueType).
		Set("category", setting.Category).
		Set("description", setting.Description).
		Set("is_public", setting.IsPublic).
		Set("updated_at", setting.UpdatedAt).
		Where(squirrel.Eq{"id": setting.ID}).
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
