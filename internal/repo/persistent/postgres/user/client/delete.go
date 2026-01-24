package client

import (
	"context"
	"time"

	apperrors "gct/pkg/errors"

	"github.com/google/uuid"
)

func (r *Repo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := r.builder.
		Update(tableName).
		Set("deleted_at", time.Now().Unix()).
		Set("updated_at", time.Now()).
		Where("id = ? AND deleted_at = 0", id).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase,
			"failed to build delete SQL query")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
