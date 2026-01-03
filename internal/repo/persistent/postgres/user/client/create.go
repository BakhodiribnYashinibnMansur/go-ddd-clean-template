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
		Columns("id", "role_id", "username", "email", "phone", "password_hash", "salt", "attributes", "active", "created_at", "updated_at", "deleted_at", "last_seen").
		Values(u.ID, u.RoleID, u.Username, u.Email, u.Phone, u.PasswordHash, u.Salt, u.Attributes, u.Active, time.Now(), time.Now(), 0, u.LastSeen).
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
