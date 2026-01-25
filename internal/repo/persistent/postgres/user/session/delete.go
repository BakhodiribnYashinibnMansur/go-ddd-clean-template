package session

import (
	"context"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Delete(ctx context.Context, filter *domain.SessionFilter) error {
	if filter == nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase,
			consts.ErrMsgInvalidInput)
	}

	query := r.builder.Delete(tableName)

	if !filter.IsIDNull() {
		query = query.Where(squirrel.Eq{schema.SessionID: *filter.ID})
	}
	if !filter.IsUserIDNull() {
		query = query.Where(squirrel.Eq{schema.SessionUserID: *filter.UserID})
	}
	if !filter.IsRevokedNull() {
		query = query.Where(squirrel.Eq{schema.SessionRevoked: *filter.Revoked})
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
