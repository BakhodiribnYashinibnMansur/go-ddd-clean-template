package client

import (
	"context"
	"time"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Update(ctx context.Context, u *domain.User) error {
	sql, args, err := r.builder.
		Update("users").
		Set("username", u.Username).
		Set("phone", u.Phone).
		Set("password_hash", u.PasswordHash).
		Set("salt", u.Salt).
		Set("updated_at", time.Now()).
		Set("last_seen", u.LastSeen).
		Where("id = ? AND deleted_at = 0", u.ID).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
			"failed to build update SQL query")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(ctx, err, "users", nil)
	}

	return nil
}
