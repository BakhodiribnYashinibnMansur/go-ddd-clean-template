package session

import (
	"context"

	"gct/internal/shared/domain/consts"
	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Delete(ctx context.Context, filter *domain.SessionFilter) error {
	if filter == nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase,
			consts.ErrMsgInvalidInput)
	}

	query := r.builder.Delete(tableName)

	if !filter.IsIDNull() {
		query = query.Where(squirrel.Eq{"id": *filter.ID})
	}
	if !filter.IsUserIDNull() {
		query = query.Where(squirrel.Eq{"user_id": *filter.UserID})
	}
	if !filter.IsRevokedNull() {
		query = query.Where(squirrel.Eq{"revoked": *filter.Revoked})
	}

	sql, args, err := query.ToSql()
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
