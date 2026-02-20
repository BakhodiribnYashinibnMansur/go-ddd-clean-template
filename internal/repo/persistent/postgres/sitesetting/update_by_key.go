package sitesetting

import (
	"context"
	"time"

	"gct/consts"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

// UpdateByKey updates a setting by its key (useful for simple value updates)
func (r *Repo) UpdateByKey(ctx context.Context, key, value string) error {
	sql, args, err := r.builder.
		Update(tableName).
		Set("value", value).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"key": key}).
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
