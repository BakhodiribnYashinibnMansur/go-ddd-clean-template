package permission

import (
	"context"
	"time"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, p *domain.Permission) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns("parent_id", "name", "created_at").
		Values(p.ParentID, p.Name, time.Now()).
		Suffix("RETURNING id, created_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build insert query")
	}

	err = r.pool.QueryRow(ctx, sql, args...).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
