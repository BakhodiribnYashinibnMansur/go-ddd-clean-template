package client

import (
	"context"
	"time"

	apperrors "gct/pkg/errors"
)

// ChangeRole updates the role_id for the user with the given ID, looking up role by name.
func (r *Repo) ChangeRole(ctx context.Context, id, role string) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET role_id = (SELECT id FROM role WHERE name = $1), updated_at = $2 WHERE id::text = $3`,
		role, now, id,
	)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
