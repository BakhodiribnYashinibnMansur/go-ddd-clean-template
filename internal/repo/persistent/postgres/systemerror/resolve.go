package systemerror

import (
	"context"
	"time"

	"gct/internal/repo/schema"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

// MarkAsResolved marks a system error as resolved
func (r *Repo) MarkAsResolved(ctx context.Context, id uuid.UUID, resolvedBy uuid.UUID) error {
	query, args, err := r.db.Builder.
		Update(schema.TableSystemError).
		Set(schema.SystemErrorIsResolved, true).
		Set(schema.SystemErrorResolvedAt, time.Now()).
		Set(schema.SystemErrorResolvedBy, resolvedBy).
		Where(squirrel.Eq{schema.SystemErrorID: id}).
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
