package role

import (
	"context"

	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Get(ctx context.Context, filter *domain.RoleFilter) (*domain.Role, error) {
	query := r.builder.Select(schema.RoleID, schema.RoleName, schema.RoleCreatedAt).From(tableName)

	if filter.ID != nil {
		query = query.Where(squirrel.Eq{schema.RoleID: *filter.ID})
	}
	if filter.Name != nil {
		query = query.Where(squirrel.Eq{schema.RoleName: *filter.Name})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build select query")
	}

	var role domain.Role
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&role.ID, &role.Name, &role.CreatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return &role, nil
}
