package policy

import (
	"context"

	"fmt"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) GetByRole(ctx context.Context, roleID uuid.UUID) ([]*domain.Policy, error) {
	// Policies can be linked to permissions, which are linked to roles.
	// Or policies can be directly linked to roles if we had that relation, but standard Domain model says Policy -> Permission.
	// Actually, policies are usually "on top of" permissions to refine them (ABAC).
	// So we need to fetch policies where permission is assigned to the role.

	sql, args, err := r.builder.
		Select(
			"p."+"id",
			"p."+"permission_id",
			"p."+"effect",
			"p."+"priority",
			"p."+"active",
			"p."+"conditions",
			"p."+"created_at",
		).
		From("policy" + " p").
		Join(fmt.Sprintf("%s rp ON p.%s = rp.%s",
			"role_permission",
			"permission_id",
			"permission_id",
		)).
		Where(squirrel.Eq{"rp." + "role_id": roleID}).
		Where(squirrel.Eq{"p." + "active": true}).
		OrderBy("p." + "priority" + " DESC"). // Higher priority first
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build select query")
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, "policy", nil)
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		var p domain.Policy
		if err := rows.Scan(&p.ID, &p.PermissionID, &p.Effect, &p.Priority, &p.Active, &p.Conditions, &p.CreatedAt); err != nil {
			return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to scan row")
		}
		policies = append(policies, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.HandlePgError(err, "policy", nil)
	}

	return policies, nil
}
