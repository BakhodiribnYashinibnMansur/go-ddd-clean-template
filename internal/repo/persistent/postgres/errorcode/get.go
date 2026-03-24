package errorcode

import (
	"context"

	"gct/internal/domain"

	"github.com/Masterminds/squirrel"
)

// GetByCode retrieves an error code by its unique code string
func (r *Repo) GetByCode(ctx context.Context, code string) (*domain.ErrorCode, error) {
	query, args, err := r.builder.
		Select(
			"id",
			"code",
			"message",
			"http_status",
			"category",
			"severity",
			"retryable",
			"retry_after",
			"suggestion",
			"created_at",
			"updated_at",
		).
		From("error_code").
		Where(squirrel.Eq{"code": code}).
		ToSql()

	if err != nil {
		r.logger.Error("failed to build get query", "error", err)
		return nil, err
	}

	var ec domain.ErrorCode
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&ec.ID,
		&ec.Code,
		&ec.Message,
		&ec.HTTPStatus,
		&ec.Category,
		&ec.Severity,
		&ec.Retryable,
		&ec.RetryAfter,
		&ec.Suggestion,
		&ec.CreatedAt,
		&ec.UpdatedAt,
	)

	if err != nil {
		if err.Error() != "no rows in result set" { // naive check
			r.logger.Error("failed to get error code", "error", err, "code", code)
		}
		return nil, err
	}

	return &ec, nil
}
