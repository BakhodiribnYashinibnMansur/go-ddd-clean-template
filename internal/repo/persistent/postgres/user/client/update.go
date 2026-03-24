package client

import (
	"context"
	"time"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) Update(ctx context.Context, u *domain.User) error {
	return pgxutil.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		sql, args, err := r.builder.
			Update(tableName).
			Set("role_id", u.RoleID).
			Set("username", u.Username).
			Set("email", u.Email).
			Set("phone", u.Phone).
			SetMap(squirrel.Eq{
				"password_hash": u.PasswordHash,
				"salt":         u.Salt,
				"attributes":   u.Attributes,
				"active":       u.Active,
				"is_approved":   u.IsApproved,
			}).
			Set("updated_at", time.Now()).
			Set("last_seen", u.LastSeen).
			Where(squirrel.Eq{"id": u.ID}).
			ToSql()
		if err != nil {
			return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build update query")
		}

		_, err = tx.Exec(ctx, sql, args...)
		if err != nil {
			return apperrors.HandlePgError(err, tableName, nil)
		}

		return nil
	})
}
