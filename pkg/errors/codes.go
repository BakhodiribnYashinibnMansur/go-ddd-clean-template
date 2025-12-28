package errors

// Error kodlar va numeric code'lar

const (
	// 400 - Bad Request
	ErrBadRequest  = "BAD_REQUEST"
	CodeBadRequest = "1001"

	ErrInvalidInput  = "INVALID_INPUT"
	CodeInvalidInput = "1002"

	ErrValidation  = "VALIDATION_ERROR"
	CodeValidation = "1003"

	// 401 - Unauthorized
	ErrUnauthorized  = "UNAUTHORIZED"
	CodeUnauthorized = "1004"

	ErrInvalidToken  = "INVALID_TOKEN"
	CodeInvalidToken = "1005"

	ErrExpiredToken  = "EXPIRED_TOKEN"
	CodeExpiredToken = "1006"

	ErrRevokedToken  = "REVOKED_TOKEN"
	CodeRevokedToken = "1007"

	// 403 - Forbidden
	ErrForbidden  = "FORBIDDEN"
	CodeForbidden = "1008"

	ErrPermissionDenied  = "PERMISSION_DENIED"
	CodePermissionDenied = "1010"

	ErrDisabledAccount  = "DISABLED_ACCOUNT"
	CodeDisabledAccount = "1011"

	// 404 - Not Found
	ErrNotFound  = "NOT_FOUND"
	CodeNotFound = "1012"

	ErrUserNotFound  = "USER_NOT_FOUND"
	CodeUserNotFound = "1013"

	ErrSessionNotFound  = "SESSION_NOT_FOUND"
	CodeSessionNotFound = "1014"

	// 409 - Conflict
	ErrConflict  = "CONFLICT"
	CodeConflict = "1015"

	ErrAlreadyExists  = "ALREADY_EXISTS"
	CodeAlreadyExists = "1016"

	// 500 - Internal Server Error
	ErrInternal  = "INTERNAL_ERROR"
	CodeInternal = "1017"

	ErrDatabase  = "DATABASE_ERROR"
	CodeDatabase = "1018"

	ErrUnknown  = "UNKNOWN_ERROR"
	CodeUnknown = "1019"

	// 504 - Timeout
	ErrTimeout  = "TIMEOUT"
	CodeTimeout = "1020"
)
