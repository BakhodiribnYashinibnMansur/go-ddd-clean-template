package client

import (
	"context"
	"time"

	apperrors "github.com/evrone/go-clean-template/pkg/errors"
)

func (r *Repo) Delete(ctx context.Context, id int64) error {
	sql, args, err := r.builder.
		Update("users").
		Set("deleted_at", time.Now().Unix()).
		Set("updated_at", time.Now()).
		Where("id = ? AND deleted_at = 0", id).
		ToSql()
	if err != nil {
		return apperrors.AutoSource(
			apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
				"failed to build delete SQL query")).
			WithField("user_id", id).
			WithDetails("Error occurred while building soft DELETE query")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		// Use centralized PostgreSQL error handler!
		return apperrors.HandlePgError(ctx, err, "users", map[string]any{
			"user_id": id,
		})
	}

	return nil
}
