package errorcode

import (
	"context"
	"time"

	"gct/internal/domain"

	"github.com/Masterminds/squirrel"
)

// UpdateErrorCodeInput represents input for updating an error code
type UpdateErrorCodeInput struct {
	Message    *string               `json:"message"`
	HTTPStatus *int                  `json:"http_status"`
	Category   *domain.ErrorCategory `json:"category"`
	Severity   *domain.ErrorSeverity `json:"severity"`
	Retryable  *bool                 `json:"retryable"`
	RetryAfter *int                  `json:"retry_after"`
	Suggestion *string               `json:"suggestion"`
}

// Update updates an existing error code
func (r *Repo) Update(ctx context.Context, code string, input UpdateErrorCodeInput) (*domain.ErrorCode, error) {
	builder := r.builder.Update("error_code").
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"code": code}).
		Suffix("RETURNING " +
			"id" + ", " +
			"code" + ", " +
			"message" + ", " +
			"http_status" + ", " +
			"category" + ", " +
			"severity" + ", " +
			"retryable" + ", " +
			"retry_after" + ", " +
			"suggestion" + ", " +
			"created_at" + ", " +
			"updated_at")

	if input.Message != nil {
		builder = builder.Set("message", *input.Message)
	}
	if input.HTTPStatus != nil {
		builder = builder.Set("http_status", *input.HTTPStatus)
	}
	if input.Category != nil {
		builder = builder.Set("category", *input.Category)
	}
	if input.Severity != nil {
		builder = builder.Set("severity", *input.Severity)
	}
	if input.Retryable != nil {
		builder = builder.Set("retryable", *input.Retryable)
	}
	if input.RetryAfter != nil {
		builder = builder.Set("retry_after", *input.RetryAfter)
	}
	if input.Suggestion != nil {
		builder = builder.Set("suggestion", *input.Suggestion)
	}

	query, args, err := builder.ToSql()
	if err != nil {
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
		r.logger.Error("failed to update error code", "error", err, "code", code)
		return nil, err
	}

	return &ec, nil
}
