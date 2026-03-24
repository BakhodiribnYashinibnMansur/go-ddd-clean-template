package filemetadata

import (
	"context"

	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
)

// Delete removes a file_metadata record by UUID string.
func (r *Repo) Delete(ctx context.Context, id string) error {
	sql, args, err := r.builder.Delete(table).Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build delete")
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, table, nil)
	}
	return nil
}
