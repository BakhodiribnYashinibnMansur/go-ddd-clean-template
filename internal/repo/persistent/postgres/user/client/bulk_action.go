package client

import (
	"context"
	"time"

	apperrors "gct/pkg/errors"
)

// BulkDeactivate sets active=false for all users with the given IDs.
func (r *Repo) BulkDeactivate(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET active = false, updated_at = $1 WHERE id::text = ANY($2)`,
		now, ids,
	)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// BulkDelete soft-deletes all users with the given IDs.
func (r *Repo) BulkDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET deleted_at = $1, updated_at = $2 WHERE id::text = ANY($3)`,
		now.Unix(), now, ids,
	)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
