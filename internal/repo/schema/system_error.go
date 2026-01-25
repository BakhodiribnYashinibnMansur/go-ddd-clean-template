package schema

// Table name
const TableSystemError = "system_errors"

// SystemError table columns
const (
	SystemErrorID          = "id"
	SystemErrorCode        = "code"
	SystemErrorMessage     = "message"
	SystemErrorStackTrace  = "stack_trace"
	SystemErrorMetadata    = "metadata"
	SystemErrorSeverity    = "severity"
	SystemErrorServiceName = "service_name"
	SystemErrorRequestID   = "request_id"
	SystemErrorUserID      = "user_id"
	SystemErrorIPAddress   = "ip_address"
	SystemErrorPath        = "path"
	SystemErrorMethod      = "method"
	SystemErrorIsResolved  = "is_resolved"
	SystemErrorResolvedAt  = "resolved_at"
	SystemErrorResolvedBy  = "resolved_by"
	SystemErrorCreatedAt   = "created_at"
)
