package policy

import (
	"context"

	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/google/uuid"
)

func (r *Repo) Toggle(ctx context.Context, id uuid.UUID) error {
	sql := "UPDATE " + tableName + " SET active = NOT active WHERE id = $1"

	tag, err := r.pool.Exec(ctx, sql, id)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	if tag.RowsAffected() == 0 {
		return apperrors.NewRepoError(apperrors.ErrRepoNotFound, "policy not found")
	}

	return nil
}
