package errorcode

import (
	"context"
	"time"

	"gct/internal/domain"
	"gct/internal/repo/schema"

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
	builder := r.db.Builder.Update(schema.TableErrorCode).
		Set(schema.ErrorCodeUpdatedAt, time.Now()).
		Where(squirrel.Eq{schema.ErrorCodeCode: code}).
		Suffix("RETURNING " +
			schema.ErrorCodeID + ", " +
			schema.ErrorCodeCode + ", " +
			schema.ErrorCodeMessage + ", " +
			schema.ErrorCodeHTTPStatus + ", " +
			schema.ErrorCodeCategory + ", " +
			schema.ErrorCodeSeverity + ", " +
			schema.ErrorCodeRetryable + ", " +
			schema.ErrorCodeRetryAfter + ", " +
			schema.ErrorCodeSuggestion + ", " +
			schema.ErrorCodeCreatedAt + ", " +
			schema.ErrorCodeUpdatedAt)

	if input.Message != nil {
		builder = builder.Set(schema.ErrorCodeMessage, *input.Message)
	}
	if input.HTTPStatus != nil {
		builder = builder.Set(schema.ErrorCodeHTTPStatus, *input.HTTPStatus)
	}
	if input.Category != nil {
		builder = builder.Set(schema.ErrorCodeCategory, *input.Category)
	}
	if input.Severity != nil {
		builder = builder.Set(schema.ErrorCodeSeverity, *input.Severity)
	}
	if input.Retryable != nil {
		builder = builder.Set(schema.ErrorCodeRetryable, *input.Retryable)
	}
	if input.RetryAfter != nil {
		builder = builder.Set(schema.ErrorCodeRetryAfter, *input.RetryAfter)
	}
	if input.Suggestion != nil {
		builder = builder.Set(schema.ErrorCodeSuggestion, *input.Suggestion)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	var ec domain.ErrorCode
	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(
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
