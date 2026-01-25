package domain

import (
	"time"

	"github.com/google/uuid"
)

type ErrorSeverity string

const (
	SeverityLow      ErrorSeverity = "LOW"
	SeverityMedium   ErrorSeverity = "MEDIUM"
	SeverityHigh     ErrorSeverity = "HIGH"
	SeverityCritical ErrorSeverity = "CRITICAL"
)

type ErrorCategory string

const (
	CategoryData       ErrorCategory = "DATA"
	CategoryAuth       ErrorCategory = "AUTH"
	CategorySystem     ErrorCategory = "SYSTEM"
	CategoryValidation ErrorCategory = "VALIDATION"
	CategoryBusiness   ErrorCategory = "BUSINESS"
	CategoryUnknown    ErrorCategory = "UNKNOWN"
)

type ErrorCode struct {
	ID         uuid.UUID     `json:"id" db:"id"`
	Code       string        `json:"code" db:"code"`
	Message    string        `json:"message" db:"message"`
	HTTPStatus int           `json:"http_status" db:"http_status"`
	Category   ErrorCategory `json:"category" db:"category"`
	Severity   ErrorSeverity `json:"severity" db:"severity"`
	Retryable  bool          `json:"retryable" db:"retryable"`
	RetryAfter int           `json:"retry_after" db:"retry_after"`
	Suggestion string        `json:"suggestion" db:"suggestion"`
	CreatedAt  time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at" db:"updated_at"`
}
