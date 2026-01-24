package role

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]*domain.Permission, error) {
	sql, args, err := r.builder.
		Select("p.id", "p.name", "p.created_at").
		From("permission p").
		Join("role_permission rp ON p.id = rp.permission_id").
		Where(squirrel.Eq{"rp.role_id": roleID}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build select query")
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, "permission", nil)
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
		return nil, apperrors.HandlePgError(err, "permission", nil)
	}

	return permissions, nil
}
