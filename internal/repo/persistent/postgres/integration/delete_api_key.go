package integration

import (
	"context"

	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) DeleteAPIKey(ctx context.Context, id uuid.UUID) error {
	sql, args, err := r.builder.
		Update(tableAPIKeys).
		Set("deleted_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": id}).
		Where(squirrel.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, tableAPIKeys, nil)
	}

	return nil
}
