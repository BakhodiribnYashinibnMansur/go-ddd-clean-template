package role

import (
	"context"

	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Update(ctx context.Context, role *domain.Role) error {
	sql, args, err := r.builder.
		Update(tableName).
		Set(schema.RoleName, role.Name).
		Where(squirrel.Eq{schema.RoleID: role.ID}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build update query")
	}

	tag, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	if tag.RowsAffected() == 0 {
		return apperrors.NewRepoError(apperrors.ErrRepoNotFound, "role not found")
	}

	return nil
}
