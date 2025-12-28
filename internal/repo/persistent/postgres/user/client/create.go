package client

import (
	"context"
	"time"

	"github.com/evrone/go-clean-template/internal/domain"
	apperrors "github.com/evrone/go-clean-template/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, u *domain.User) error {
	sql, args, err := r.builder.
		Insert("users").
		Columns("username", "phone", "password_hash", "salt", "created_at", "updated_at", "deleted_at", "last_seen").
		Values(u.Username, u.Phone, u.PasswordHash, u.Salt, time.Now(), time.Now(), 0, u.LastSeen).
		ToSql()
	if err != nil {
		return apperrors.AutoSource(
			apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
				"failed to build insert SQL query")).
			WithField("operation", "build_insert_query").
			WithField("username", u.Username).
			WithField("phone", u.Phone).
			WithDetails("Error occurred while building INSERT query")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		// Use centralized PostgreSQL error handler!
		return apperrors.HandlePgError(ctx, err, "users", map[string]any{
			"username": u.Username,
			"phone":    u.Phone,
		})
	}

	return nil
}
