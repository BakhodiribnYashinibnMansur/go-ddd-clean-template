package session

import (
	"context"

	"github.com/Masterminds/squirrel"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Delete(ctx context.Context, filter *domain.SessionFilter) error {
	sql, args, err := r.builder.
		Delete("session").
		Where(squirrel.Eq{"id": *filter.ID}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
			"failed to build delete SQL query")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(ctx, err, "session", nil)
	}

	return nil
}
