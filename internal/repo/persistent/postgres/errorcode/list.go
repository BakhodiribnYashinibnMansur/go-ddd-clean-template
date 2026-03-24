package errorcode

import (
	"context"

	"gct/internal/domain"
)

// List returns all error codes
func (r *Repo) List(ctx context.Context) ([]*domain.ErrorCode, error) {
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
		OrderBy("code" + " ASC").
		ToSql()

	if err != nil {
		r.logger.Error("failed to build list query", "error", err)
		return nil, err
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to list error codes", "error", err)
		return nil, err
	}
	defer rows.Close()

	var codes []*domain.ErrorCode
	for rows.Next() {
		var ec domain.ErrorCode
		err := rows.Scan(
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
			return nil, err
		}
		codes = append(codes, &ec)
	}

	return codes, nil
}
