package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// ErrorCode is the aggregate root representing a cataloged application error.
// It serves as a central registry entry that maps a machine-readable code to human-readable
// metadata (message, suggestion) and operational hints (httpStatus, retryable, retryAfter).
// The code field is immutable after creation — only the descriptive fields can be updated.
type ErrorCode struct {
	shared.AggregateRoot
	code       string
	message    string
	messageUz  string
	messageRu  string
	httpStatus int
	category   string
	severity   string
	retryable  bool
	retryAfter int
	suggestion string
}

// NewErrorCode creates a new ErrorCode aggregate and raises an ErrorCodeCreated event.
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
	ec.AddEvent(NewErrorCodeCreated(ec.ID(), code, message, httpStatus))
	return ec
}

// SetTranslations sets the Uzbek and Russian message translations.
func (ec *ErrorCode) SetTranslations(uz, ru string) {
	ec.messageUz = uz
	ec.messageRu = ru
}

// ReconstructErrorCode rebuilds an ErrorCode aggregate from persisted data.
func ReconstructErrorCode(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	code, message, messageUz, messageRu string,
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
		messageUz:     messageUz,
		messageRu:     messageRu,
		httpStatus:    httpStatus,
		category:      category,
		severity:      severity,
		retryable:     retryable,
		retryAfter:    retryAfter,
		suggestion:    suggestion,
	}
}

// Update replaces all mutable fields of the error code and raises an ErrorCodeUpdated event.
// Note: the code field itself is immutable — only metadata fields are updated.
func (ec *ErrorCode) Update(
	message, messageUz, messageRu string,
	httpStatus int,
	category, severity string,
	retryable bool,
	retryAfter int,
	suggestion string,
) {
	ec.message = message
	ec.messageUz = messageUz
	ec.messageRu = messageRu
	ec.httpStatus = httpStatus
	ec.category = category
	ec.severity = severity
	ec.retryable = retryable
	ec.retryAfter = retryAfter
	ec.suggestion = suggestion
	ec.Touch()
	ec.AddEvent(NewErrorCodeUpdated(ec.ID(), ec.code, message, httpStatus))
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (ec *ErrorCode) Code() string       { return ec.code }
func (ec *ErrorCode) Message() string    { return ec.message }
func (ec *ErrorCode) MessageUz() string  { return ec.messageUz }
func (ec *ErrorCode) MessageRu() string  { return ec.messageRu }
func (ec *ErrorCode) HTTPStatus() int    { return ec.httpStatus }
func (ec *ErrorCode) Category() string   { return ec.category }
func (ec *ErrorCode) Severity() string   { return ec.severity }
func (ec *ErrorCode) Retryable() bool    { return ec.retryable }
func (ec *ErrorCode) RetryAfter() int    { return ec.retryAfter }
func (ec *ErrorCode) Suggestion() string { return ec.suggestion }
