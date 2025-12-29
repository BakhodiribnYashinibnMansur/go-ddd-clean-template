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
	case ErrBadRequest, ErrInvalidInput, ErrValidation:
		return 400

	// 401 errors
	case ErrUnauthorized, ErrInvalidToken, ErrExpiredToken, ErrRevokedToken:
		return 401

	// 403 errors
	case ErrForbidden, ErrPermissionDenied, ErrDisabledAccount:
		return 403

	// 404 errors
	case ErrNotFound, ErrUserNotFound, ErrSessionNotFound:
		return 404

	// 409 errors
	case ErrConflict, ErrAlreadyExists:
		return 409

	// 500 errors
	case ErrInternal, ErrDatabase, ErrUnknown:
		return 500

	// 504 errors
	case ErrTimeout:
		return 504

	default:
		return 500
	}
}

// getUserMessage returns user-facing message by error code
func getUserMessage(code string) string {
	if msg := getCommonUserMessage(code); msg != "" {
		return msg
	}
	if msg := getAuthUserMessage(code); msg != "" {
		return msg
	}
	if msg := getResourceUserMessage(code); msg != "" {
		return msg
	}
	if msg := getSystemUserMessage(code); msg != "" {
		return msg
	}
	return "An error occurred"
}

func getCommonUserMessage(code string) string {
	switch code {
	case ErrBadRequest:
		return "Bad request"
	case ErrInvalidInput:
		return "Invalid input provided"
	case ErrValidation:
		return "Validation failed"
	default:
		return ""
	}
}

func getAuthUserMessage(code string) string {
	switch code {
	case ErrUnauthorized:
		return "Authentication required"
	case ErrInvalidToken:
		return "Invalid token"
	case ErrExpiredToken:
		return "Token has expired"
	case ErrRevokedToken:
		return "Token has been revoked"
	case ErrForbidden:
		return "Access denied"
	case ErrPermissionDenied:
		return "You don't have permission to perform this action"
	case ErrDisabledAccount:
		return "Account is disabled"
	default:
		return ""
	}
}

func getResourceUserMessage(code string) string {
	switch code {
	case ErrNotFound:
		return "Not found"
	case ErrUserNotFound:
		return "User not found"
	case ErrSessionNotFound:
		return "Session not found"
	case ErrConflict:
		return "Resource already exists"
	case ErrAlreadyExists:
		return "Already exists"
	default:
		return ""
	}
}

func getSystemUserMessage(code string) string {
	switch code {
	case ErrInternal:
		return "Internal error"
	case ErrDatabase:
		return "Database error"
	case ErrTimeout:
		return "Request timeout"
	case ErrUnknown:
		return "Unknown error"
	default:
		return ""
	}
}

// getNumericCode returns numeric code by error type
func getNumericCode(code string) string {
	if c := getRepoNumericCode(code); c != "" {
		return c
	}
	if c := getServiceNumericCode(code); c != "" {
		return c
	}
	if c := getHandlerNumericCode(code); c != "" {
		return c
	}
	if c := getLegacyNumericCode(code); c != "" {
		return c
	}
	return "9999"
}

func getRepoNumericCode(code string) string {
	return repoNumericCodes[code]
}

func getServiceNumericCode(code string) string {
	return serviceNumericCodes[code]
}

func getHandlerNumericCode(code string) string {
	return handlerNumericCodes[code]
}

func getLegacyNumericCode(code string) string {
	return legacyNumericCodes[code]
}

var (
	repoNumericCodes = map[string]string{
		ErrRepoDatabase:      CodeRepoDatabase,
		ErrRepoNotFound:      CodeRepoNotFound,
		ErrRepoAlreadyExists: CodeRepoAlreadyExists,
		ErrRepoConstraint:    CodeRepoConstraint,
		ErrRepoUnknown:       CodeRepoUnknown,
	}

	serviceNumericCodes = map[string]string{
		ErrServiceInvalidInput:  CodeServiceInvalidInput,
		ErrServiceValidation:    CodeServiceValidation,
		ErrServiceNotFound:      CodeServiceNotFound,
		ErrServiceAlreadyExists: CodeServiceAlreadyExists,
		ErrServiceUnauthorized:  CodeServiceUnauthorized,
		ErrServiceForbidden:     CodeServiceForbidden,
		ErrServiceConflict:      CodeServiceConflict,
		ErrServiceBusinessRule:  CodeServiceBusinessRule,
		ErrServiceDependency:    CodeServiceDependency,
		ErrServiceUnknown:       CodeServiceUnknown,
	}

	handlerNumericCodes = map[string]string{
		ErrHandlerBadRequest:   CodeHandlerBadRequest,
		ErrHandlerUnauthorized: CodeHandlerUnauthorized,
		ErrHandlerForbidden:    CodeHandlerForbidden,
		ErrHandlerNotFound:     CodeHandlerNotFound,
		ErrHandlerConflict:     CodeHandlerConflict,
		ErrHandlerInternal:     CodeHandlerInternal,
		ErrHandlerUnknown:      CodeHandlerUnknown,
	}

	legacyNumericCodes = map[string]string{
		ErrBadRequest:       CodeBadRequest,
		ErrInvalidInput:     CodeInvalidInput,
		ErrValidation:       CodeValidation,
		ErrUnauthorized:     CodeUnauthorized,
		ErrInvalidToken:     CodeInvalidToken,
		ErrExpiredToken:     CodeExpiredToken,
		ErrRevokedToken:     CodeRevokedToken,
		ErrForbidden:        CodeForbidden,
		ErrPermissionDenied: CodePermissionDenied,
		ErrDisabledAccount:  CodeDisabledAccount,
		ErrNotFound:         CodeNotFound,
		ErrUserNotFound:     CodeUserNotFound,
		ErrSessionNotFound:  CodeSessionNotFound,
		ErrConflict:         CodeConflict,
		ErrAlreadyExists:    CodeAlreadyExists,
		ErrInternal:         CodeInternal,
		ErrDatabase:         CodeDatabase,
		ErrTimeout:          CodeTimeout,
		ErrUnknown:          CodeUnknown,
	}
)
