package errors

import (
	"errors"
)

// ============================================================================
// Handler Layer Error Codes
// ============================================================================

const (
	// HTTP handler errors
	ErrHandlerBadRequest  = "HANDLER_BAD_REQUEST"
	CodeHandlerBadRequest = "4000"

	ErrHandlerUnauthorized  = "HANDLER_UNAUTHORIZED"
	CodeHandlerUnauthorized = "4001"

	ErrHandlerForbidden  = "HANDLER_FORBIDDEN"
	CodeHandlerForbidden = "4003"

	ErrHandlerNotFound  = "HANDLER_NOT_FOUND"
	CodeHandlerNotFound = "4004"

	ErrHandlerMethodNotAllowed  = "HANDLER_METHOD_NOT_ALLOWED"
	CodeHandlerMethodNotAllowed = "4005"

	ErrHandlerConflict  = "HANDLER_CONFLICT"
	CodeHandlerConflict = "4009"

	ErrHandlerTooManyRequests  = "HANDLER_TOO_MANY_REQUESTS"
	CodeHandlerTooManyRequests = "4029"

	ErrHandlerInternal  = "HANDLER_INTERNAL_ERROR"
	CodeHandlerInternal = "5000"

	ErrHandlerUnknown  = "HANDLER_UNKNOWN_ERROR"
	CodeHandlerUnknown = "5099"
)

// Handler error messages
var handlerMessages = map[string]string{
	ErrHandlerBadRequest:       "Bad request",
	ErrHandlerUnauthorized:     "Unauthorized access",
	ErrHandlerForbidden:        "Forbidden access",
	ErrHandlerNotFound:         "Resource not found",
	ErrHandlerMethodNotAllowed: "Method not allowed",
	ErrHandlerConflict:         "Resource conflict",
	ErrHandlerTooManyRequests:  "Too many requests",
	ErrHandlerInternal:         "Internal server error",
	ErrHandlerUnknown:          "Unknown error",
}

// NewHandlerError creates a new handler error
func NewHandlerError(code, message string) *AppError {
	return New(code, message)
}

// WrapHandlerError wraps an error as handler error
func WrapHandlerError(err error, code, message string) *AppError {
	return Wrap(err, code, message)
}

// MapServiceToHandlerError maps service error to handler error
func MapServiceToHandlerError(err error) *AppError {
	if err == nil {
		return nil
	}

	// If it's already our AppError, check the code
	var appErr *AppError
	if errors.As(err, &appErr) {
		switch appErr.Type {
		case ErrServiceNotFound:
			return NewHandlerError(ErrHandlerNotFound, "Resource not found").
				WithDetails(appErr.Message)

		case ErrServiceInvalidInput, ErrServiceValidation:
			return NewHandlerError(ErrHandlerBadRequest, "Invalid request").
				WithDetails(appErr.Message)

		case ErrServiceUnauthorized:
			return NewHandlerError(ErrHandlerUnauthorized, "Unauthorized").
				WithDetails(appErr.Message)

		case ErrServiceForbidden:
			return NewHandlerError(ErrHandlerForbidden, "Forbidden").
				WithDetails(appErr.Message)

		case ErrServiceConflict, ErrServiceAlreadyExists:
			return NewHandlerError(ErrHandlerConflict, "Resource conflict").
				WithDetails(appErr.Message)

		default:
			return WrapHandlerError(err, ErrHandlerInternal, "Internal server error")
		}
	}

	// For non-AppError, wrap as internal error
	return WrapHandlerError(err, ErrHandlerInternal, "Internal server error")
}

// MapToHTTPStatus maps error code to HTTP status code
func MapToHTTPStatus(code string) int {
	switch code {
	// 400 errors
	case ErrHandlerBadRequest, ErrServiceInvalidInput, ErrServiceValidation:
		return 400

	// 401 errors
	case ErrHandlerUnauthorized, ErrServiceUnauthorized:
		return 401

	// 403 errors
	case ErrHandlerForbidden, ErrServiceForbidden:
		return 403

	// 404 errors
	case ErrHandlerNotFound, ErrServiceNotFound, ErrRepoNotFound:
		return 404

	// 405 errors
	case ErrHandlerMethodNotAllowed:
		return 405

	// 409 errors
	case ErrHandlerConflict, ErrServiceConflict, ErrServiceAlreadyExists, ErrRepoAlreadyExists:
		return 409

	// 500 errors
	case ErrHandlerInternal, ErrServiceUnknown, ErrRepoDatabase, ErrRepoUnknown:
		return 500

	// 504 errors
	case ErrRepoTimeout:
		return 504

	default:
		return 500
	}
}
