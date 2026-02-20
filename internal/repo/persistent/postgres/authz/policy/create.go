package policy

import (
	"context"
	"fmt"
	"time"

	"gct/consts"
	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, p *domain.Policy) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	sql, args, err := r.builder.
		Insert(tableName).
		Columns(
			"permission_id",
			"effect",
			"priority",
			"active",
			"conditions",
			"created_at",
		).
		Values(p.PermissionID, p.Effect, p.Priority, p.Active, p.Conditions, time.Now()).
		Suffix(fmt.Sprintf("RETURNING %s, %s", "id", "created_at")).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	err = tx.QueryRow(ctx, sql, args...).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	if err := tx.Commit(ctx); err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to commit transaction")
	}

	return nil
}
