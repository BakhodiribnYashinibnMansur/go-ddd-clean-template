package consts

// Error codes for the SystemError aggregate. Each code is persisted in the system_errors table
// and used for filtering, alerting, and dashboard aggregation. Add new codes here rather than
// using ad-hoc strings to maintain a searchable, finite error taxonomy.
const (
	// Authentication & Authorization Errors
	ErrCodeAuthFailed              = "AUTH_FAILED"
	ErrCodeInvalidToken            = "INVALID_TOKEN"
	ErrCodeTokenExpired            = "TOKEN_EXPIRED"
	ErrCodeTokenMissing            = "TOKEN_MISSING"
	ErrCodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"
	ErrCodeInvalidCredentials      = "INVALID_CREDENTIALS"
	ErrCodeSessionExpired          = "SESSION_EXPIRED"
	ErrCodeSessionNotFound         = "SESSION_NOT_FOUND"

	// User Errors
	ErrCodeUserNotFound      = "USER_NOT_FOUND"
	ErrCodeUserAlreadyExists = "USER_ALREADY_EXISTS"
	ErrCodeUserNotApproved   = "USER_NOT_APPROVED"
	ErrCodeUserBlocked       = "USER_BLOCKED"
	ErrCodeUserDeleted       = "USER_DELETED"
	ErrCodeUserInactive      = "USER_INACTIVE"

	// Database Errors
	ErrCodeDatabaseError       = "DATABASE_ERROR"
	ErrCodeQueryTimeout        = "QUERY_TIMEOUT"
	ErrCodeConnectionLost      = "CONNECTION_LOST"
	ErrCodeTransactionFailed   = "TRANSACTION_FAILED"
	ErrCodeDeadlock            = "DEADLOCK"
	ErrCodeConstraintViolation = "CONSTRAINT_VIOLATION"

	// Validation Errors
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeMissingField     = "MISSING_FIELD"
	ErrCodeInvalidFormat    = "INVALID_FORMAT"
	ErrCodeInvalidEmail     = "INVALID_EMAIL"
	ErrCodeInvalidPhone     = "INVALID_PHONE"
	ErrCodePasswordTooWeak  = "PASSWORD_TOO_WEAK"

	// External Service Errors
	ErrCodeExternalServiceError = "EXTERNAL_SERVICE_ERROR"
	ErrCodeAPITimeout           = "API_TIMEOUT"
	ErrCodeAPIRateLimited       = "API_RATE_LIMITED"
	ErrCodeAPIUnavailable       = "API_UNAVAILABLE"
	ErrCodeMyIDError            = "MYID_ERROR"
	ErrCodeMyIDTimeout          = "MYID_TIMEOUT"

	// File & Storage Errors
	ErrCodeFileNotFound     = "FILE_NOT_FOUND"
	ErrCodeFileUploadFailed = "FILE_UPLOAD_FAILED"
	ErrCodeFileDeleteFailed = "FILE_DELETE_FAILED"
	ErrCodeInvalidFileType  = "INVALID_FILE_TYPE"
	ErrCodeFileTooLarge     = "FILE_TOO_LARGE"
	ErrCodeStorageError     = "STORAGE_ERROR"

	// Cache Errors
	ErrCodeCacheError      = "CACHE_ERROR"
	ErrCodeCacheMiss       = "CACHE_MISS"
	ErrCodeCacheSetFailed  = "CACHE_SET_FAILED"
	ErrCodeCacheInvalidate = "CACHE_INVALIDATE_FAILED"

	// Business Logic Errors
	ErrCodeInvalidOperation = "INVALID_OPERATION"
	ErrCodeOperationFailed  = "OPERATION_FAILED"
	ErrCodeResourceNotFound = "RESOURCE_NOT_FOUND"
	ErrCodeResourceLocked   = "RESOURCE_LOCKED"
	ErrCodeDuplicateEntry   = "DUPLICATE_ENTRY"
	ErrCodeInvalidState     = "INVALID_STATE"
	ErrCodeMethodNotAllowed = "METHOD_NOT_ALLOWED"

	// System Errors
	ErrCodeInternalError       = "INTERNAL_ERROR"
	ErrCodeConfigError         = "CONFIG_ERROR"
	ErrCodeInitializationError = "INITIALIZATION_ERROR"
	ErrCodeShutdownError       = "SHUTDOWN_ERROR"
	ErrCodePanic               = "PANIC"

	// Rate Limiting & Throttling
	ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	ErrCodeTooManyRequests   = "TOO_MANY_REQUESTS"
	ErrCodeQuotaExceeded     = "QUOTA_EXCEEDED"

	// Audit & Compliance
	ErrCodeAuditLogFailed      = "AUDIT_LOG_FAILED"
	ErrCodeComplianceViolation = "COMPLIANCE_VIOLATION"
	ErrCodeUnauthorizedAccess  = "UNAUTHORIZED_ACCESS"

	// Migration & Data Transfer
	ErrCodeMigrationFailed    = "MIGRATION_FAILED"
	ErrCodeDataTransferFailed = "DATA_TRANSFER_FAILED"
	ErrCodeDataCorruption     = "DATA_CORRUPTION"
)

// Error severity levels for SystemError. FATAL and PANIC trigger immediate alerting (e.g., Telegram notifications).
const (
	SeverityWarn  = "WARN"
	SeverityError = "ERROR"
	SeverityFatal = "FATAL"
	SeverityPanic = "PANIC"
)
