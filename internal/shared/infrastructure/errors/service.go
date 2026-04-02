package errors

import (
	"errors"
	"strings"

	shared "gct/internal/shared/domain"
)

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

	// Authz specific errors
	ErrServiceRoleNotFound  = "SERVICE_ROLE_NOT_FOUND"
	CodeServiceRoleNotFound = "3010"

	ErrServicePermissionNotFound  = "SERVICE_PERMISSION_NOT_FOUND"
	CodeServicePermissionNotFound = "3011"

	ErrServicePolicyViolation  = "SERVICE_POLICY_VIOLATION"
	CodeServicePolicyViolation = "3012"

	ErrServiceScopeNotFound  = "SERVICE_SCOPE_NOT_FOUND"
	CodeServiceScopeNotFound = "3013"

	ErrServiceRelationNotFound  = "SERVICE_RELATION_NOT_FOUND"
	CodeServiceRelationNotFound = "3014"

	ErrServiceUnknown  = "SERVICE_UNKNOWN_ERROR"
	CodeServiceUnknown = "3099"
)

// Service error messages
var serviceMessages = map[string]string{
	ErrServiceInvalidInput:       "Invalid input provided",
	ErrServiceValidation:         "Validation failed",
	ErrServiceNotFound:           "Resource not found",
	ErrServiceAlreadyExists:      "Resource already exists",
	ErrServiceUnauthorized:       "Authentication required",
	ErrServiceForbidden:          "Permission denied",
	ErrServiceConflict:           "Resource conflict",
	ErrServiceBusinessRule:       "Business rule violation",
	ErrServiceDependency:         "Dependency service error",
	ErrServiceUnknown:            "Unknown service error",
	ErrServiceRoleNotFound:       "Role not found",
	ErrServicePermissionNotFound: "Permission not found",
	ErrServicePolicyViolation:    "Policy violation",
	ErrServiceScopeNotFound:      "Scope not found",
	ErrServiceRelationNotFound:   "Relation not found",
}

// NewServiceError creates a new service error
func NewServiceError(code, message string) *AppError {
	return New(code, message)
}

// WrapServiceError wraps an error as service error
func WrapServiceError(err error, code, message string) *AppError {
	return Wrap(err, code, message)
}

// MapDomainToServiceError converts a DomainError to a service-layer AppError.
// The original domain code is preserved as the Type so that handlers/clients
// can identify the exact business error (e.g. USER_INACTIVE, USER_WEAK_PASSWORD).
func MapDomainToServiceError(err error) *AppError {
	var domErr *shared.DomainError
	if !errors.As(err, &domErr) {
		return nil
	}

	code := domErr.Code()

	return &AppError{
		Type:       code,
		Code:       GetNumericCode(code),
		Message:    domErr.Error(),
		HTTPStatus: getHTTPStatus(code),
		UserMsg:    getUserMessage(code),
		Severity:   GetSeverity(code),
		Category:   GetCategory(code),
		Err:        err,
		Stack:      captureStack(),
	}
}

// MapToServiceError converts any error to a service-layer AppError.
// Handles DomainError, repo AppError, external AppError, and raw errors.
func MapToServiceError(err error) error {
	if err == nil {
		return nil
	}

	// 1. DomainError → AppError preserving domain code
	if domAppErr := MapDomainToServiceError(err); domAppErr != nil {
		return domAppErr
	}

	// 2. Already an AppError
	var appErr *AppError
	if errors.As(err, &appErr) {
		// Service/handler/external errors pass through
		if strings.HasPrefix(appErr.Type, "SERVICE_") ||
			strings.HasPrefix(appErr.Type, "HANDLER_") ||
			strings.HasPrefix(appErr.Type, "EXT_") {
			return err
		}
		// Repo error → map to service error
		if strings.HasPrefix(appErr.Type, "REPO_") {
			return MapRepoToServiceError(err)
		}
		// Other AppError (legacy codes) → pass through
		return err
	}

	// 3. Raw error → unknown service error
	return WrapServiceError(err, ErrServiceUnknown, "")
}

// MapRepoToServiceError maps repository error to service error
func MapRepoToServiceError(err error, notFoundCode ...string) *AppError {
	if err == nil {
		return nil
	}

	// If it's already our AppError, check the code
	var appErr *AppError
	if errors.As(err, &appErr) {
		switch appErr.Type {
		case ErrRepoNotFound:
			code := ErrServiceNotFound
			msg := "Resource not found"
			if len(notFoundCode) > 0 {
				code = notFoundCode[0]
				if m, ok := serviceMessages[code]; ok {
					msg = m
				}
			}
			return NewServiceError(code, msg).
				WithDetails(appErr.Message)
		case ErrRepoAlreadyExists:
			return NewServiceError(ErrServiceAlreadyExists, "Resource already exists").
				WithDetails(appErr.Message)
		case ErrRepoConstraint:
			return NewServiceError(ErrServiceConflict, "Resource conflict").
				WithDetails(appErr.Message)
		default:
			return WrapServiceError(err, ErrServiceDependency, "Repository error")
		}
	}

	// For non-AppError, wrap as unknown
	return WrapServiceError(err, ErrServiceUnknown, "Unknown error occurred")
}
