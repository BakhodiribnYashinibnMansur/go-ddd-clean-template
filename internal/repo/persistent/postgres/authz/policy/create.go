package policy

import (
	"context"
	"fmt"
	"time"

	"gct/internal/shared/domain/consts"
	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/jackc/pgx/v5"
)

func (r *Repo) Create(ctx context.Context, p *domain.Policy) error {
	return pgxutil.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
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

		return nil
	})
}
