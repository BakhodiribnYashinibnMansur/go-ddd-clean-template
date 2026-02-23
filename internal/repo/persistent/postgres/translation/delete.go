package translation

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Delete(ctx context.Context, filter domain.TranslationFilter) error {
	query := r.builder.
		Delete(tableName).
		Where(squirrel.Eq{
			"entity_type": filter.EntityType,
			"entity_id":   filter.EntityID,
		})

	if filter.LangCode != nil {
		query = query.Where(squirrel.Eq{"lang_code": *filter.LangCode})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build delete query")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
