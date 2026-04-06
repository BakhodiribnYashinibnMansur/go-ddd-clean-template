package errorx

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

	ErrHandlerNotImplemented  = "HANDLER_NOT_IMPLEMENTED"
	CodeHandlerNotImplemented = "5001"

	ErrHandlerServiceUnavailable  = "HANDLER_SERVICE_UNAVAILABLE"
	CodeHandlerServiceUnavailable = "5003"

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
	ErrHandlerTooManyRequests:    "Too many requests",
	ErrHandlerNotImplemented:    "Not implemented",
	ErrHandlerServiceUnavailable: "Service unavailable",
	ErrHandlerInternal:          "Internal server error",
	ErrHandlerUnknown:           "Unknown error",
}

// NewHandlerError creates a new handler error
func NewHandlerError(code, message string) *AppError {
	return New(code, message)
}

// WrapHandlerError wraps an error as handler error
func WrapHandlerError(err error, code, message string) *AppError {
	return Wrap(err, code, message)
}

// MapServiceToHandlerError maps service error to handler error.
// Domain codes (6xxx) and external codes (7xxx) are preserved as-is
// so that clients receive the specific error code. Repo/service codes
// are translated to generic handler codes to hide infrastructure details.
func MapServiceToHandlerError(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		// Domain codes (from MapDomainToServiceError) — preserve as-is.
		// These have registered HTTP statuses in domain_codes.go.
		if status := GetHTTPStatus(appErr.Type); status != 0 {
			appErr.HTTPStatus = status
			return appErr
		}

		// Service codes → handler codes (hide infra details)
		switch appErr.Type {
		case ErrServiceNotFound, ErrServiceRoleNotFound,
			ErrServicePermissionNotFound, ErrServiceScopeNotFound,
			ErrServiceRelationNotFound:
			return NewHandlerError(ErrHandlerNotFound, "").
				WithDetails(appErr.Message)

		case ErrServiceInvalidInput, ErrServiceValidation:
			return NewHandlerError(ErrHandlerBadRequest, "").
				WithDetails(appErr.Message)

		case ErrServiceUnauthorized:
			return NewHandlerError(ErrHandlerUnauthorized, "").
				WithDetails(appErr.Message)

		case ErrServiceForbidden, ErrServicePolicyViolation:
			return NewHandlerError(ErrHandlerForbidden, "").
				WithDetails(appErr.Message)

		case ErrServiceConflict, ErrServiceAlreadyExists:
			return NewHandlerError(ErrHandlerConflict, "").
				WithDetails(appErr.Message)

		case ErrServiceBusinessRule:
			return NewHandlerError(ErrHandlerBadRequest, "").
				WithDetails(appErr.Message)

		case ErrServiceDependency:
			return NewHandlerError(ErrHandlerServiceUnavailable, "").
				WithDetails(appErr.Message)

		default:
			return WrapHandlerError(err, ErrHandlerInternal, "")
		}
	}

	return WrapHandlerError(err, ErrHandlerInternal, "")
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

	// 429 errors
	case ErrHandlerTooManyRequests:
		return 429

	// 500 errors
	case ErrHandlerInternal, ErrServiceUnknown, ErrRepoDatabase, ErrRepoUnknown:
		return 500

	// 501 errors
	case ErrHandlerNotImplemented:
		return 501

	// 503 errors
	case ErrHandlerServiceUnavailable:
		return 503

	// 504 errors
	case ErrRepoTimeout:
		return 504

	default:
		return 500
	}
}
