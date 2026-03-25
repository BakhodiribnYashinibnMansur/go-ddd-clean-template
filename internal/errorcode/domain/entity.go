package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// ErrorCode is the aggregate root for application error codes.
type ErrorCode struct {
	shared.AggregateRoot
	code       string
	message    string
	httpStatus int
	category   string
	severity   string
	retryable  bool
	retryAfter int
	suggestion string
}

// NewErrorCode creates a new ErrorCode aggregate and raises an ErrorCodeUpdated event.
func NewErrorCode(
	code, message string,
	httpStatus int,
	category, severity string,
	retryable bool,
	retryAfter int,
	suggestion string,
) *ErrorCode {
	ec := &ErrorCode{
		AggregateRoot: shared.NewAggregateRoot(),
		code:          code,
		message:       message,
		httpStatus:    httpStatus,
		category:      category,
		severity:      severity,
		retryable:     retryable,
		retryAfter:    retryAfter,
		suggestion:    suggestion,
	}
	ec.AddEvent(NewErrorCodeUpdated(ec.ID(), code, message))
	return ec
}

// ReconstructErrorCode rebuilds an ErrorCode aggregate from persisted data.
func ReconstructErrorCode(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	code, message string,
	httpStatus int,
	category, severity string,
	retryable bool,
	retryAfter int,
	suggestion string,
) *ErrorCode {
	return &ErrorCode{
		AggregateRoot: shared.NewAggregateRootWithID(id, createdAt, updatedAt, nil),
		code:          code,
		message:       message,
		httpStatus:    httpStatus,
		category:      category,
		severity:      severity,
		retryable:     retryable,
		retryAfter:    retryAfter,
		suggestion:    suggestion,
	}
}

// Update modifies the error code fields.
func (ec *ErrorCode) Update(
	message string,
	httpStatus int,
	category, severity string,
	retryable bool,
	retryAfter int,
	suggestion string,
) {
	ec.message = message
	ec.httpStatus = httpStatus
	ec.category = category
	ec.severity = severity
	ec.retryable = retryable
	ec.retryAfter = retryAfter
	ec.suggestion = suggestion
	ec.Touch()
	ec.AddEvent(NewErrorCodeUpdated(ec.ID(), ec.code, message))
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (ec *ErrorCode) Code() string       { return ec.code }
func (ec *ErrorCode) Message() string    { return ec.message }
func (ec *ErrorCode) HTTPStatus() int    { return ec.httpStatus }
func (ec *ErrorCode) Category() string   { return ec.category }
func (ec *ErrorCode) Severity() string   { return ec.severity }
func (ec *ErrorCode) Retryable() bool    { return ec.retryable }
func (ec *ErrorCode) RetryAfter() int    { return ec.retryAfter }
func (ec *ErrorCode) Suggestion() string { return ec.suggestion }
