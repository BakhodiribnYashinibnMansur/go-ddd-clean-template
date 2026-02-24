package systemerror

import (
	"context"
	"time"

	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) Resolve(ctx context.Context, id uuid.UUID, resolvedBy *uuid.UUID) error {
	now := time.Now()
	sql, args, err := r.builder.
		Update(tableName).
		Set("is_resolved", true).
		Set("resolved_at", now).
		Set("resolved_by", resolvedBy).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build resolve update")
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return apperrors.HandlePgError(err, tableName, nil)
}
