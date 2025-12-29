package errors

import "context"

// ============================================================================
// Service Layer Error Codes
// ============================================================================

const (
	// Business logic errors
	ErrServiceInvalidInput  = "SERVICE_INVALID_INPUT"
	CodeServiceInvalidInput = "3001"

	ErrServiceValidation  = "SERVICE_VALIDATION_ERROR"
	CodeServiceValidation = "3002"

	ErrServiceNotFound  = "SERVICE_NOT_FOUND"
	CodeServiceNotFound = "3003"

	ErrServiceAlreadyExists  = "SERVICE_ALREADY_EXISTS"
	CodeServiceAlreadyExists = "3004"

	ErrServiceUnauthorized  = "SERVICE_UNAUTHORIZED"
	CodeServiceUnauthorized = "3005"

	ErrServiceForbidden  = "SERVICE_FORBIDDEN"
	CodeServiceForbidden = "3006"

	ErrServiceConflict  = "SERVICE_CONFLICT"
	CodeServiceConflict = "3007"

	ErrServiceBusinessRule  = "SERVICE_BUSINESS_RULE_VIOLATION"
	CodeServiceBusinessRule = "3008"

	ErrServiceDependency  = "SERVICE_DEPENDENCY_ERROR"
	CodeServiceDependency = "3009"

	ErrServiceUnknown  = "SERVICE_UNKNOWN_ERROR"
	CodeServiceUnknown = "3099"
)

// Service error messages
var serviceMessages = map[string]string{
	ErrServiceInvalidInput:  "Invalid input provided",
	ErrServiceValidation:    "Validation failed",
	ErrServiceNotFound:      "Resource not found",
	ErrServiceAlreadyExists: "Resource already exists",
	ErrServiceUnauthorized:  "Authentication required",
	ErrServiceForbidden:     "Permission denied",
	ErrServiceConflict:      "Resource conflict",
	ErrServiceBusinessRule:  "Business rule violation",
	ErrServiceDependency:    "Dependency service error",
	ErrServiceUnknown:       "Unknown service error",
}

// NewServiceError creates a new service error
func NewServiceError(ctx context.Context, code, message string) *AppError {
	return New(ctx, code, message)
}

// WrapServiceError wraps an error as service error
func WrapServiceError(ctx context.Context, err error, code, message string) *AppError {
	return Wrap(ctx, err, code, message)
}

// MapRepoToServiceError maps repository error to service error
func MapRepoToServiceError(ctx context.Context, err error) *AppError {
	if err == nil {
		return nil
	}

	// If it's already our AppError, check the code
	if appErr, ok := err.(*AppError); ok {
		switch appErr.Type {
		case ErrRepoNotFound:
			return NewServiceError(ctx, ErrServiceNotFound, "Resource not found").
				WithDetails(appErr.Message)
		case ErrRepoAlreadyExists:
			return NewServiceError(ctx, ErrServiceAlreadyExists, "Resource already exists").
				WithDetails(appErr.Message)
		case ErrRepoConstraint:
			return NewServiceError(ctx, ErrServiceConflict, "Resource conflict").
				WithDetails(appErr.Message)
		default:
			return WrapServiceError(ctx, err, ErrServiceDependency, "Repository error")
		}
	}

	// For non-AppError, wrap as unknown
	return WrapServiceError(ctx, err, ErrServiceUnknown, "Unknown error occurred")
}
