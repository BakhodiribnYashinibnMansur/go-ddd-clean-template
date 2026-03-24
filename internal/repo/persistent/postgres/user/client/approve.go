package client

import (
	"context"
	"time"

	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
)

// Approve sets is_approved=true for the user with the given ID.
func (r *Repo) Approve(ctx context.Context, id string) error {
	sql, args, err := r.builder.
		Update(tableName).
		Set("is_approved", true).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build approve query")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
