package setting

import (
	"context"

	"gct/consts"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) Delete(ctx context.Context, userID uuid.UUID, key string) error {
	sql, args, err := r.builder.
		Delete(tableName).
		Where(squirrel.Eq{"user_id": userID, "key": key}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	return apperrors.HandlePgError(err, tableName, nil)
}
