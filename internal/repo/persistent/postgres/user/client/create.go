package client

import (
	"context"
	"time"

	"gct/internal/shared/domain/consts"
	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/jackc/pgx/v5"
)

func (r *Repo) Create(ctx context.Context, u *domain.User) error {
	return pgxutil.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		sql, args, err := r.builder.
			Insert(tableName).
			Columns(
				"id",
				"role_id",
				"username",
				"email",
				"phone",
				"password_hash",
				"salt",
				"attributes",
				"active",
				"is_approved",
				"created_at",
				"updated_at",
				"deleted_at",
				"last_seen",
			).
			Values(u.ID, u.RoleID, u.Username, u.Email, u.Phone, u.PasswordHash, u.Salt,
				u.Attributes, u.Active, u.IsApproved, time.Now(), time.Now(), 0, u.LastSeen).
			ToSql()
		if err != nil {
			return apperrors.NewRepoError(apperrors.ErrRepoDatabase,
				consts.ErrMsgFailedToBuildInsert)
		}

		_, err = tx.Exec(ctx, sql, args...)
		if err != nil {
			return apperrors.HandlePgError(err, tableName, nil)
		}

		return nil
	})
}
