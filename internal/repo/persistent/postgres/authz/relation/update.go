package relation

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Update(ctx context.Context, relation *domain.Relation) error {
	sql, args, err := r.builder.
		Update(tableName).
		Set("name", relation.Name).
		Set("type", relation.Type).
		Where(squirrel.Eq{"id": relation.ID}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build update query")
	}

	tag, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	if tag.RowsAffected() == 0 {
		return apperrors.NewRepoError(apperrors.ErrRepoNotFound, "relation not found")
	}

	return nil
}
