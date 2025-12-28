package client

import (
	"context"
	"time"

	"github.com/evrone/go-clean-template/internal/domain"
	apperrors "github.com/evrone/go-clean-template/pkg/errors"
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
		return apperrors.AutoSource(
			apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
				"failed to build update SQL query")).
			WithField("user_id", u.ID).
			WithDetails("Error occurred while building UPDATE query")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		// Use centralized PostgreSQL error handler!
		return apperrors.HandlePgError(ctx, err, "users", map[string]any{
			"user_id": u.ID,
			"phone":   u.Phone,
		})
	}

	return nil
}
