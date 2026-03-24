package errorcode

import (
	"context"
	"gct/internal/domain"
)

// CreateErrorCodeInput represents input for creating an error code
type CreateErrorCodeInput struct {
	Code       string               `json:"code" binding:"required"`
	Message    string               `json:"message" binding:"required"`
	HTTPStatus int                  `json:"http_status" binding:"required"`
	Category   domain.ErrorCategory `json:"category"`
	Severity   domain.ErrorSeverity `json:"severity"`
	Retryable  bool                 `json:"retryable"`
	RetryAfter int                  `json:"retry_after"`
	Suggestion string               `json:"suggestion"`
}

// Create inserts a new error code into the database
func (r *Repo) Create(ctx context.Context, input CreateErrorCodeInput) (*domain.ErrorCode, error) {
	query, args, err := r.builder.
		Insert("error_code").
		Columns(
			"code",
			"message",
			"http_status",
			"category",
			"severity",
			"retryable",
			"retry_after",
			"suggestion",
		).
		Values(
			input.Code,
			input.Message,
			input.HTTPStatus,
			input.Category,
			input.Severity,
			input.Retryable,
			input.RetryAfter,
			input.Suggestion,
		).
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
			"updated_at").
		ToSql()

	if err != nil {
		r.logger.Error("failed to build create query", "error", err)
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
		r.logger.Error("failed to create error code", "error", err, "code", input.Code)
		return nil, err
	}

	return &ec, nil
}
