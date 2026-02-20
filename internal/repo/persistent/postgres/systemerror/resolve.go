package systemerror

import (
	"context"
	"time"


	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

// MarkAsResolved marks a system error as resolved
func (r *Repo) MarkAsResolved(ctx context.Context, id uuid.UUID, resolvedBy uuid.UUID) error {
	query, args, err := r.db.Builder.
		Update("system_errors").
		Set("is_resolved", true).
		Set("resolved_at", time.Now()).
		Set("resolved_by", resolvedBy).
		Where(squirrel.Eq{"id": id}).
		ToSql()

	if err != nil {
		r.logger.Error("failed to build mark resolved query", "error", err)
		return err
	}

	_, err = r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to mark system error as resolved", "error", err, "id", id)
		return err
	}

	r.logger.Info("system error marked as resolved", "error_id", id, "resolved_by", resolvedBy)
	return nil
}
