package errorcode

import (
	"context"

	"gct/internal/domain"
	"gct/internal/repo/schema"
)

// List returns all error codes
func (r *Repo) List(ctx context.Context) ([]*domain.ErrorCode, error) {
	query, args, err := r.db.Builder.
		Select(
			schema.ErrorCodeID,
			schema.ErrorCodeCode,
			schema.ErrorCodeMessage,
			schema.ErrorCodeHTTPStatus,
			schema.ErrorCodeCategory,
			schema.ErrorCodeSeverity,
			schema.ErrorCodeRetryable,
			schema.ErrorCodeRetryAfter,
			schema.ErrorCodeSuggestion,
			schema.ErrorCodeCreatedAt,
			schema.ErrorCodeUpdatedAt,
		).
		From(schema.TableErrorCode).
		OrderBy(schema.ErrorCodeCode + " ASC").
		ToSql()

	if err != nil {
		r.logger.Error("failed to build list query", "error", err)
		return nil, err
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
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
