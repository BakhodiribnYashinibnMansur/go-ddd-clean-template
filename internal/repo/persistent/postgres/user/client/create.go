package client

import (
	"context"
	"time"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, u *domain.User) error {
	sql, args, err := r.builder.
		Insert("users").
		Columns("username", "phone", "password_hash", "salt", "created_at", "updated_at", "deleted_at", "last_seen").
		Values(u.Username, u.Phone, u.PasswordHash, u.Salt, time.Now(), time.Now(), 0, u.LastSeen).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
			"failed to build insert SQL query")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(ctx, err, "users", nil)
	}

	return nil
}
