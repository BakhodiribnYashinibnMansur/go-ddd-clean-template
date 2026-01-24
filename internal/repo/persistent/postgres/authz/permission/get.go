package permission

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Get(ctx context.Context, filter *domain.PermissionFilter) (*domain.Permission, error) {
	query := r.builder.Select("id", "parent_id", "name", "created_at").From(tableName)

	if filter.ID != nil {
		query = query.Where(squirrel.Eq{"id": *filter.ID})
	}
	if filter.Name != nil {
		query = query.Where(squirrel.Eq{"name": *filter.Name})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build select query")
	}

	var p domain.Permission
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&p.ID, &p.ParentID, &p.Name, &p.CreatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return &p, nil
}
