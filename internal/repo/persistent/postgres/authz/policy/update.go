package policy

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) Update(ctx context.Context, p *domain.Policy) error {
	return pgxutil.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		sql, args, err := r.builder.
			Update(tableName).
			Set("permission_id", p.PermissionID).
			Set("effect", p.Effect).
			Set("priority", p.Priority).
			Set("active", p.Active).
			Set("conditions", p.Conditions).
			Where(squirrel.Eq{"id": p.ID}).
			ToSql()
		if err != nil {
			return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build update query")
		}

		tag, err := tx.Exec(ctx, sql, args...)
		if err != nil {
			return apperrors.HandlePgError(err, tableName, nil)
		}

		if tag.RowsAffected() == 0 {
			return apperrors.NewRepoError(apperrors.ErrRepoNotFound, "policy not found")
		}

		return nil
	})
}
