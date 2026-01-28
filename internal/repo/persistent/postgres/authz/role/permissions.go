package role

import (
	"context"
	"time"

	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) AddPermission(ctx context.Context, roleID, permID uuid.UUID) error {
	sql, args, err := r.builder.
		Insert(schema.TableRolePermission).
		Columns(schema.RolePermissionRoleID, schema.RolePermissionPermissionID, schema.RolePermissionCreatedAt).
		Values(roleID, permID, time.Now()).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build insert query")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, schema.TableRolePermission, nil)
	}

	return nil
}

func (r *Repo) RemovePermission(ctx context.Context, roleID, permID uuid.UUID) error {
	sql, args, err := r.builder.
		Delete(schema.TableRolePermission).
		Where(squirrel.Eq{schema.RolePermissionRoleID: roleID}).
		Where(squirrel.Eq{schema.RolePermissionPermissionID: permID}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build delete query")
	}

	tag, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, schema.TableRolePermission, nil)
	}

	if tag.RowsAffected() == 0 {
		return apperrors.NewRepoError(apperrors.ErrRepoNotFound, "role permission not found")
	}

	return nil
}
