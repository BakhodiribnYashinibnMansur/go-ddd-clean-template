package schema

// Table name
const TableErrorCode = "error_code"

// ErrorCode table columns
const (
	ErrorCodeID         = "id"
	ErrorCodeCode       = "code"
	ErrorCodeMessage    = "message"
	ErrorCodeHTTPStatus = "http_status"
	ErrorCodeCategory   = "category"
	ErrorCodeSeverity   = "severity"
	ErrorCodeRetryable  = "retryable"
	ErrorCodeRetryAfter = "retry_after"
	ErrorCodeSuggestion = "suggestion"
	ErrorCodeCreatedAt  = "created_at"
	ErrorCodeUpdatedAt  = "updated_at"
)
