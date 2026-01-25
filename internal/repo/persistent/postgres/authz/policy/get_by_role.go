package policy

import (
	"context"

	"fmt"

	"gct/internal/domain"
	"gct/internal/repo/schema"
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
			"p."+schema.PolicyID,
			"p."+schema.PolicyPermissionID,
			"p."+schema.PolicyEffect,
			"p."+schema.PolicyPriority,
			"p."+schema.PolicyActive,
			"p."+schema.PolicyConditions,
			"p."+schema.PolicyCreatedAt,
		).
		From(schema.TablePolicy + " p").
		Join(fmt.Sprintf("%s rp ON p.%s = rp.%s",
			schema.TableRolePermission,
			schema.PolicyPermissionID,
			schema.RolePermissionPermissionID,
		)).
		Where(squirrel.Eq{"rp." + schema.RolePermissionRoleID: roleID}).
		Where(squirrel.Eq{"p." + schema.PolicyActive: true}).
		OrderBy("p." + schema.PolicyPriority + " DESC"). // Higher priority first
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build select query")
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, schema.TablePolicy, nil)
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
		return nil, apperrors.HandlePgError(err, schema.TablePolicy, nil)
	}

	return policies, nil
}
