package errors

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"runtime"
)

// AppError custom error structure.
//
// Security note: fields tagged `json:"-"` MUST NEVER be serialized to API
// responses. They may contain internal details, wrapped infrastructure errors,
// stack traces, or raw request/response payloads (which can include passwords,
// tokens, or PII). These fields are for internal logging and debugging only.
// The HTTP layer builds its own response struct from the safe fields
// (Type, Code, HTTPStatus, UserMsg, Details, Severity, Category, Suggestion).
type AppError struct {
	Type       string         // Error type (e.g.: "USER_NOT_FOUND")
	Code       string         // Numeric error code (e.g.: "4041")
	Message    string         `json:"-"` // Developer message (internal)
	HTTPStatus int            // HTTP status code
	UserMsg    string         // User-facing message
	Details    string         // Detailed explanation
	Severity   ErrorSeverity  // Error severity
	Category   ErrorCategory  // Error category
	Suggestion string         // Help suggestion
	Fields     map[string]any `json:"-"` // Additional data (may contain sensitive values)
	Err        error          `json:"-"` // Wrapped error (may expose infra internals)
	Stack      []uintptr      `json:"-"` // Stack trace (never expose to clients)
	Input      any            `json:"-"` // Input data (may contain credentials/PII)
	Output     any            `json:"-"` // Output data (may contain credentials/PII)
}

// Error interface implementation
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithField adds a field
func (e *AppError) WithField(key string, value any) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]any)
	}
	e.Fields[key] = value
	return e
}

// WithInput adds input data to error
func (e *AppError) WithInput(input any) *AppError {
	e.Input = input
	return e
}

// WithOutput adds output data to error
func (e *AppError) WithOutput(output any) *AppError {
	e.Output = output
	return e
}

// WithDetails adds detailed information
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithSuggestion adds a suggestion
func (e *AppError) WithSuggestion(suggestion string) *AppError {
	e.Suggestion = suggestion
	return e
}

// Fingerprint returns a short hash identifying this error type + location.
// Useful for grouping identical errors.
func (e *AppError) Fingerprint() string {
	src := e.Type
	if len(e.Stack) > 0 {
		frames := runtime.CallersFrames(e.Stack[:1])
		if f, ok := frames.Next(); ok {
			src += f.Function
		}
	}
	h := sha256.Sum256([]byte(src))
	return hex.EncodeToString(h[:8])
}

// New creates new error. If message is empty, it is resolved from the registry.
func New(code, message string) *AppError {
	userMsg := getUserMessage(code)
	if userMsg == "" && message != "" {
		userMsg = message
	}
	if userMsg == "" {
		userMsg = "An error occurred"
	}

	return &AppError{
		Type:       code,
		Code:       GetNumericCode(code),
		Message:    message,
		HTTPStatus: getHTTPStatus(code),
		UserMsg:    userMsg,
		Severity:   GetSeverity(code),
		Category:   GetCategory(code),
		Stack:      captureStack(),
	}
}

// Wrap wraps an existing error. If message is empty, it is resolved from the registry.
func Wrap(err error, code, message string) *AppError {
	if err == nil {
		return nil
	}

	userMsg := getUserMessage(code)
	if userMsg == "" && message != "" {
		userMsg = message
	}
	if userMsg == "" {
		userMsg = "An error occurred"
	}

	return &AppError{
		Type:       code,
		Code:       GetNumericCode(code),
		Message:    message,
		HTTPStatus: getHTTPStatus(code),
		UserMsg:    userMsg,
		Severity:   GetSeverity(code),
		Category:   GetCategory(code),
		Err:        err,
		Stack:      captureStack(),
	}
}

// Is checks if error matches the code
func Is(err error, code string) bool {
	var e *AppError
	if errors.As(err, &e) {
		return e.Type == code
	}
	return false
}

// GetCode returns error code from error
func GetCode(err error) string {
	var e *AppError
	if errors.As(err, &e) {
		return e.Type
	}
	return ""
}

// captureStack creates a stack trace
func captureStack() []uintptr {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	return pcs[0:n]
}

// getHTTPStatus returns HTTP status by error code
func getHTTPStatus(code string) int {
	// 1. Check dynamic configuration first (DB loaded)
	if status := GetHTTPStatus(code); status != 0 {
		return status
	}

	// 2. Fallback to hardcoded defaults
	switch code {
	// 400 errors
	case ErrBadRequest, ErrInvalidInput, ErrValidation, ErrServiceInvalidInput, ErrServiceValidation, ErrHandlerBadRequest:
		return 400

	// 401 errors
	case ErrUnauthorized, ErrInvalidToken, ErrExpiredToken, ErrRevokedToken, ErrServiceUnauthorized, ErrHandlerUnauthorized:
		return 401

	// 403 errors
	case ErrForbidden, ErrPermissionDenied, ErrDisabledAccount, ErrServiceForbidden, ErrServicePolicyViolation, ErrHandlerForbidden:
		return 403

	// 404 errors
	case ErrNotFound, ErrUserNotFound, ErrSessionNotFound, ErrServiceNotFound, ErrServiceRoleNotFound, ErrServicePermissionNotFound, ErrServiceScopeNotFound, ErrBucketNotFound, ErrFileNotFound, ErrRepoNotFound, ErrHandlerNotFound:
		return 404

	// 409 errors
	case ErrConflict, ErrAlreadyExists, ErrServiceAlreadyExists, ErrServiceConflict, ErrRepoAlreadyExists, ErrRepoConstraint, ErrHandlerConflict:
		return 409

	// 500 errors
	case ErrInternal, ErrDatabase, ErrUnknown, ErrServiceUnknown, ErrServiceDependency, ErrRepoDatabase, ErrRepoConnection, ErrRepoTransaction, ErrRepoUnknown, ErrHandlerInternal, ErrHandlerUnknown:
		return 500

	// 504 errors
	case ErrTimeout, ErrRepoTimeout:
		return 504

	// 429 errors
	case ErrHandlerTooManyRequests:
		return 429

	// 501 errors
	case ErrHandlerNotImplemented:
		return 501

	// 503 errors
	case ErrHandlerServiceUnavailable:
		return 503

	default:
		return 500
	}
}
