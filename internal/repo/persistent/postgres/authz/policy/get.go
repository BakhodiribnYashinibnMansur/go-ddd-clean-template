package policy

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
	"github.com/Masterminds/squirrel"
)

func (r *Repo) Get(ctx context.Context, filter *domain.PolicyFilter) (*domain.Policy, error) {
	query := r.builder.Select("id", "permission_id", "effect", "priority", "active", "conditions", "created_at").From("policy")

	if filter.ID != nil {
		query = query.Where(squirrel.Eq{"id": *filter.ID})
	}
	if filter.PermissionID != nil {
		query = query.Where(squirrel.Eq{"permission_id": *filter.PermissionID})
	}
	if filter.Active != nil {
		query = query.Where(squirrel.Eq{"active": *filter.Active})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build select query")
	}

	var p domain.Policy
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&p.ID, &p.PermissionID, &p.Effect, &p.Priority, &p.Active, &p.Conditions, &p.CreatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(ctx, err, "policy", nil)
	}

	return &p, nil
}
