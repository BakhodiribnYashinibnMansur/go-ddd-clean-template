package policy

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Update(ctx context.Context, p *domain.Policy) error {
	sql, args, err := r.builder.
		Update("policy").
		Set("permission_id", p.PermissionID).
		Set("effect", p.Effect).
		Set("priority", p.Priority).
		Set("active", p.Active).
		Set("conditions", p.Conditions).
		Where(squirrel.Eq{"id": p.ID}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build update query")
	}

	tag, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(ctx, err, "policy", nil)
	}

	if tag.RowsAffected() == 0 {
		return apperrors.NewRepoError(ctx, apperrors.ErrRepoNotFound, "policy not found")
	}

	return nil
}
