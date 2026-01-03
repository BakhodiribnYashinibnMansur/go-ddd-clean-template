package errors

import (
	"context"
	"errors"
	"fmt"
	"runtime"
)

// AppError custom error structure
type AppError struct {
	Type       string         // Error type (e.g.: "USER_NOT_FOUND")
	Code       string         // Numeric error code (e.g.: "4041")
	Message    string         // Developer message
	HTTPStatus int            // HTTP status code
	UserMsg    string         // User-facing message
	Details    string         // Detailed explanation
	Fields     map[string]any // Additional data
	Err        error          // Wrapped error
	Stack      []uintptr      // Stack trace
	Input      any            // Input data
	Output     any            // Output data
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

// New creates new error
func New(ctx context.Context, code, message string) *AppError {
	return &AppError{
		Type:       code,
		Code:       getNumericCode(code),
		Message:    message,
		HTTPStatus: getHTTPStatus(code),
		UserMsg:    getUserMessage(code),
		Stack:      captureStack(),
	}
}

// Wrap wraps an existing error
func Wrap(ctx context.Context, err error, code, message string) *AppError {
	if err == nil {
		return nil
	}

	return &AppError{
		Type:       code,
		Code:       getNumericCode(code),
		Message:    message,
		HTTPStatus: getHTTPStatus(code),
		UserMsg:    getUserMessage(code),
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

	default:
		return 500
	}
}
