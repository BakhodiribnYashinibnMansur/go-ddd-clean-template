package relation

import (
	"context"
	"time"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, relation *domain.Relation) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns("type", "name", "created_at").
		Values(relation.Type, relation.Name, time.Now()).
		Suffix("RETURNING id, created_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build insert SQL query")
	}

	err = r.pool.QueryRow(ctx, sql, args...).Scan(&relation.ID, &relation.CreatedAt)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
