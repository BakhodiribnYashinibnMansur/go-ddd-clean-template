package role

import (
	"context"
	"time"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, role *domain.Role) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns("name", "created_at").
		Values(role.Name, time.Now()).
		Suffix("RETURNING id, created_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build insert SQL query")
	}

	err = r.pool.QueryRow(ctx, sql, args...).Scan(&role.ID, &role.CreatedAt)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
