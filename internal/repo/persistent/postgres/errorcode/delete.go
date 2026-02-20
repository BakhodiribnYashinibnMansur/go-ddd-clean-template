package errorcode

import (
	"context"


	"github.com/Masterminds/squirrel"
)

// Delete removes an error code
func (r *Repo) Delete(ctx context.Context, code string) error {
	query, args, err := r.db.Builder.
		Delete("error_code").
		Where(squirrel.Eq{"code": code}).
		ToSql()

	if err != nil {
		r.logger.Error("failed to build delete query", "error", err)
		return err
	}

	_, err = r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to delete error code", "error", err, "code", code)
		return err
	}
	return nil
}
