package application

import (
	"time"

	"gct/internal/context/admin/supporting/errorcode/domain"
)

// ErrorCodeView is a read-model DTO returned by query handlers.
type ErrorCodeView struct {
	ID         domain.ErrorCodeID `json:"id"`
	Code       string             `json:"code"`
	Message    string             `json:"message"`
	MessageUz  string             `json:"message_uz"`
	MessageRu  string             `json:"message_ru"`
	HTTPStatus int                `json:"http_status"`
	Category   string             `json:"category"`
	Severity   string             `json:"severity"`
	Retryable  bool               `json:"retryable"`
	RetryAfter int                `json:"retry_after"`
	Suggestion string             `json:"suggestion"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}
