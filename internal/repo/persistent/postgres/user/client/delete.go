package client

import (
	"context"
	"time"

	apperrors "gct/pkg/errors"
)

func (r *Repo) Delete(ctx context.Context, id int64) error {
	sql, args, err := r.builder.
		Update("users").
		Set("deleted_at", time.Now().Unix()).
		Set("updated_at", time.Now()).
		Where("id = ? AND deleted_at = 0", id).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
			"failed to build delete SQL query")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(ctx, err, "users", nil)
	}

	return nil
}
