package role

import (
	"context"

	"fmt"

	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]*domain.Permission, error) {
	sql, args, err := r.builder.
		Select(
			"p."+schema.PermissionID,
			"p."+schema.PermissionName,
			"p."+schema.PermissionCreatedAt,
		).
		From(schema.TablePermission + " p").
		Join(fmt.Sprintf("%s rp ON p.%s = rp.%s",
			schema.TableRolePermission,
			schema.PermissionID,
			schema.RolePermissionPermissionID,
		)).
		Where(squirrel.Eq{"rp." + schema.RolePermissionRoleID: roleID}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build select query")
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, schema.TablePermission, nil)
	}
	defer rows.Close()

	var permissions []*domain.Permission
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.CreatedAt); err != nil {
			return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to scan row")
		}
		permissions = append(permissions, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.HandlePgError(err, schema.TablePermission, nil)
	}

	return permissions, nil
}
