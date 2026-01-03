package policy

import (
	"context"
	"time"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, p *domain.Policy) error {
	sql, args, err := r.builder.
		Insert("policy").
		Columns("permission_id", "effect", "priority", "active", "conditions", "created_at").
		Values(p.PermissionID, p.Effect, p.Priority, p.Active, p.Conditions, time.Now()).
		Suffix("RETURNING id, created_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build insert query")
	}

	err = r.pool.QueryRow(ctx, sql, args...).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return apperrors.HandlePgError(ctx, err, "policy", nil)
	}

	return nil
}
