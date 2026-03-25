package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ErrorCodeFilter carries filtering parameters for listing error codes.
type ErrorCodeFilter struct {
	Code     *string
	Category *string
	Severity *string
	Limit    int64
	Offset   int64
}

// ErrorCodeRepository is the write-side repository for the ErrorCode aggregate.
type ErrorCodeRepository interface {
	Save(ctx context.Context, entity *ErrorCode) error
	Update(ctx context.Context, entity *ErrorCode) error
	FindByID(ctx context.Context, id uuid.UUID) (*ErrorCode, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// ErrorCodeView is a read-model DTO for error codes.
type ErrorCodeView struct {
	ID         uuid.UUID `json:"id"`
	Code       string    `json:"code"`
	Message    string    `json:"message"`
	HTTPStatus int       `json:"http_status"`
	Category   string    `json:"category"`
	Severity   string    `json:"severity"`
	Retryable  bool      `json:"retryable"`
	RetryAfter int       `json:"retry_after"`
	Suggestion string    `json:"suggestion"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ErrorCodeReadRepository is the read-side repository returning projected views.
type ErrorCodeReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*ErrorCodeView, error)
	List(ctx context.Context, filter ErrorCodeFilter) ([]*ErrorCodeView, int64, error)
}
