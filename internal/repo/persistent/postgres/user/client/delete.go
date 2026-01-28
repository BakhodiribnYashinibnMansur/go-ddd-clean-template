package client

import (
	"context"
	"time"

	"gct/consts"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"

	"github.com/google/uuid"
)

func (r *Repo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := r.builder.
		Update(tableName).
		Set(schema.UsersDeletedAt, time.Now().Unix()).
		Set(schema.UsersUpdatedAt, time.Now()).
		Where(schema.UsersID+" = ? AND "+schema.UsersDeletedAt+" = 0", id).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase,
			consts.ErrMsgFailedToBuildDelete)
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
