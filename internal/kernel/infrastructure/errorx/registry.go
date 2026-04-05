package errorx

import "strings"

// ErrorDefinition represents a single error definition in the system
// swagger:model ErrorDefinition
type ErrorDefinition struct {
	Code        string `json:"code"`
	NumericCode string `json:"numeric_code"`
	Message     string `json:"message"`
	Layer       string `json:"layer"`    // Repository, Service, Handler
	Category    string `json:"category"` // Data, Validation, Security, System, Business
	HTTPStatus  int    `json:"http_status,omitempty"`
}

// GetAllErrors returns a list of all defined errors in the system
//
//nolint:funlen // static data table: aggregates all error definitions across layers
func GetAllErrors() []ErrorDefinition {
	// Initialize with capacity for known errors
	defs := make([]ErrorDefinition, 0, 30)

	// ========================================================================
	// General / Common Errors
	// ========================================================================
	defs = append(defs,
		// Validation
		ErrorDefinition{Code: ErrBadRequest, NumericCode: CodeBadRequest, Message: getUserMessage(ErrBadRequest), Layer: "General", Category: "Validation", HTTPStatus: 400},
		ErrorDefinition{Code: ErrInvalidInput, NumericCode: CodeInvalidInput, Message: getUserMessage(ErrInvalidInput), Layer: "General", Category: "Validation", HTTPStatus: 400},
		ErrorDefinition{Code: ErrValidation, NumericCode: CodeValidation, Message: getUserMessage(ErrValidation), Layer: "General", Category: "Validation", HTTPStatus: 400},

		// Security
		ErrorDefinition{Code: ErrUnauthorized, NumericCode: CodeUnauthorized, Message: getUserMessage(ErrUnauthorized), Layer: "General", Category: "Security", HTTPStatus: 401},
		ErrorDefinition{Code: ErrInvalidToken, NumericCode: CodeInvalidToken, Message: getUserMessage(ErrInvalidToken), Layer: "General", Category: "Security", HTTPStatus: 401},
		ErrorDefinition{Code: ErrExpiredToken, NumericCode: CodeExpiredToken, Message: getUserMessage(ErrExpiredToken), Layer: "General", Category: "Security", HTTPStatus: 401},
		ErrorDefinition{Code: ErrRevokedToken, NumericCode: CodeRevokedToken, Message: getUserMessage(ErrRevokedToken), Layer: "General", Category: "Security", HTTPStatus: 401},
		ErrorDefinition{Code: ErrForbidden, NumericCode: CodeForbidden, Message: getUserMessage(ErrForbidden), Layer: "General", Category: "Security", HTTPStatus: 403},
		ErrorDefinition{Code: ErrPermissionDenied, NumericCode: CodePermissionDenied, Message: getUserMessage(ErrPermissionDenied), Layer: "General", Category: "Security", HTTPStatus: 403},
		ErrorDefinition{Code: ErrDisabledAccount, NumericCode: CodeDisabledAccount, Message: getUserMessage(ErrDisabledAccount), Layer: "General", Category: "Security", HTTPStatus: 403},

		// Data
		ErrorDefinition{Code: ErrNotFound, NumericCode: CodeNotFound, Message: getUserMessage(ErrNotFound), Layer: "General", Category: "Data", HTTPStatus: 404},
		ErrorDefinition{Code: ErrUserNotFound, NumericCode: CodeUserNotFound, Message: getUserMessage(ErrUserNotFound), Layer: "General", Category: "Data", HTTPStatus: 404},
		ErrorDefinition{Code: ErrSessionNotFound, NumericCode: CodeSessionNotFound, Message: getUserMessage(ErrSessionNotFound), Layer: "General", Category: "Data", HTTPStatus: 404},
		ErrorDefinition{Code: ErrConflict, NumericCode: CodeConflict, Message: getUserMessage(ErrConflict), Layer: "General", Category: "Data", HTTPStatus: 409},
		ErrorDefinition{Code: ErrAlreadyExists, NumericCode: CodeAlreadyExists, Message: getUserMessage(ErrAlreadyExists), Layer: "General", Category: "Data", HTTPStatus: 409},

		// System
		ErrorDefinition{Code: ErrInternal, NumericCode: CodeInternal, Message: getUserMessage(ErrInternal), Layer: "General", Category: "System", HTTPStatus: 500},
		ErrorDefinition{Code: ErrDatabase, NumericCode: CodeDatabase, Message: getUserMessage(ErrDatabase), Layer: "General", Category: "System", HTTPStatus: 500},
		ErrorDefinition{Code: ErrTimeout, NumericCode: CodeTimeout, Message: getUserMessage(ErrTimeout), Layer: "General", Category: "System", HTTPStatus: 504},
		ErrorDefinition{Code: ErrUnknown, NumericCode: CodeUnknown, Message: getUserMessage(ErrUnknown), Layer: "General", Category: "System", HTTPStatus: 500},

		// Storage (System/Data)
		ErrorDefinition{Code: ErrBucketNotFound, NumericCode: CodeBucketNotFound, Message: "Bucket not found", Layer: "General", Category: "System", HTTPStatus: 404},
		ErrorDefinition{Code: ErrFileNotFound, NumericCode: CodeFileNotFound, Message: "File not found", Layer: "General", Category: "Data", HTTPStatus: 404},
	)

	// ========================================================================
	// Repository Layer Errors
	// ========================================================================
	defs = append(defs,
		ErrorDefinition{Code: ErrRepoNotFound, NumericCode: CodeRepoNotFound, Message: repoMessages[ErrRepoNotFound], Layer: "Repository", Category: "Data", HTTPStatus: 404},
		ErrorDefinition{Code: ErrRepoAlreadyExists, NumericCode: CodeRepoAlreadyExists, Message: repoMessages[ErrRepoAlreadyExists], Layer: "Repository", Category: "Data", HTTPStatus: 409},
		ErrorDefinition{Code: ErrRepoDatabase, NumericCode: CodeRepoDatabase, Message: repoMessages[ErrRepoDatabase], Layer: "Repository", Category: "System", HTTPStatus: 500},
		ErrorDefinition{Code: ErrRepoTimeout, NumericCode: CodeRepoTimeout, Message: repoMessages[ErrRepoTimeout], Layer: "Repository", Category: "System", HTTPStatus: 504},
		ErrorDefinition{Code: ErrRepoConnection, NumericCode: CodeRepoConnection, Message: repoMessages[ErrRepoConnection], Layer: "Repository", Category: "System", HTTPStatus: 500},
		ErrorDefinition{Code: ErrRepoTransaction, NumericCode: CodeRepoTransaction, Message: repoMessages[ErrRepoTransaction], Layer: "Repository", Category: "System", HTTPStatus: 500},
		ErrorDefinition{Code: ErrRepoConstraint, NumericCode: CodeRepoConstraint, Message: repoMessages[ErrRepoConstraint], Layer: "Repository", Category: "Data", HTTPStatus: 400},
		ErrorDefinition{Code: ErrRepoUnknown, NumericCode: CodeRepoUnknown, Message: repoMessages[ErrRepoUnknown], Layer: "Repository", Category: "System", HTTPStatus: 500},
	)

	// ========================================================================
	// Service Layer Errors
	// ========================================================================
	defs = append(defs,
		ErrorDefinition{Code: ErrServiceInvalidInput, NumericCode: CodeServiceInvalidInput, Message: serviceMessages[ErrServiceInvalidInput], Layer: "Service", Category: "Validation", HTTPStatus: 400},
		ErrorDefinition{Code: ErrServiceValidation, NumericCode: CodeServiceValidation, Message: serviceMessages[ErrServiceValidation], Layer: "Service", Category: "Validation", HTTPStatus: 400},
		ErrorDefinition{Code: ErrServiceNotFound, NumericCode: CodeServiceNotFound, Message: serviceMessages[ErrServiceNotFound], Layer: "Service", Category: "Data", HTTPStatus: 404},
		ErrorDefinition{Code: ErrServiceAlreadyExists, NumericCode: CodeServiceAlreadyExists, Message: serviceMessages[ErrServiceAlreadyExists], Layer: "Service", Category: "Data", HTTPStatus: 409},
		ErrorDefinition{Code: ErrServiceUnauthorized, NumericCode: CodeServiceUnauthorized, Message: serviceMessages[ErrServiceUnauthorized], Layer: "Service", Category: "Security", HTTPStatus: 401},
		ErrorDefinition{Code: ErrServiceForbidden, NumericCode: CodeServiceForbidden, Message: serviceMessages[ErrServiceForbidden], Layer: "Service", Category: "Security", HTTPStatus: 403},
		ErrorDefinition{Code: ErrServiceConflict, NumericCode: CodeServiceConflict, Message: serviceMessages[ErrServiceConflict], Layer: "Service", Category: "Business", HTTPStatus: 409},
		ErrorDefinition{Code: ErrServiceBusinessRule, NumericCode: CodeServiceBusinessRule, Message: serviceMessages[ErrServiceBusinessRule], Layer: "Service", Category: "Business", HTTPStatus: 422},
		ErrorDefinition{Code: ErrServiceDependency, NumericCode: CodeServiceDependency, Message: serviceMessages[ErrServiceDependency], Layer: "Service", Category: "System", HTTPStatus: 502},
		ErrorDefinition{Code: ErrServiceRoleNotFound, NumericCode: CodeServiceRoleNotFound, Message: serviceMessages[ErrServiceRoleNotFound], Layer: "Service", Category: "Security", HTTPStatus: 404},
		ErrorDefinition{Code: ErrServicePermissionNotFound, NumericCode: CodeServicePermissionNotFound, Message: serviceMessages[ErrServicePermissionNotFound], Layer: "Service", Category: "Security", HTTPStatus: 404},
		ErrorDefinition{Code: ErrServicePolicyViolation, NumericCode: CodeServicePolicyViolation, Message: serviceMessages[ErrServicePolicyViolation], Layer: "Service", Category: "Security", HTTPStatus: 403},
		ErrorDefinition{Code: ErrServiceScopeNotFound, NumericCode: CodeServiceScopeNotFound, Message: serviceMessages[ErrServiceScopeNotFound], Layer: "Service", Category: "Security", HTTPStatus: 404},
		ErrorDefinition{Code: ErrServiceUnknown, NumericCode: CodeServiceUnknown, Message: serviceMessages[ErrServiceUnknown], Layer: "Service", Category: "System", HTTPStatus: 500},
	)

	// ========================================================================
	// Handler Layer Errors
	// ========================================================================
	defs = append(defs,
		ErrorDefinition{Code: ErrHandlerBadRequest, NumericCode: CodeHandlerBadRequest, Message: handlerMessages[ErrHandlerBadRequest], Layer: "Handler", Category: "Validation", HTTPStatus: 400},
		ErrorDefinition{Code: ErrHandlerUnauthorized, NumericCode: CodeHandlerUnauthorized, Message: handlerMessages[ErrHandlerUnauthorized], Layer: "Handler", Category: "Security", HTTPStatus: 401},
		ErrorDefinition{Code: ErrHandlerForbidden, NumericCode: CodeHandlerForbidden, Message: handlerMessages[ErrHandlerForbidden], Layer: "Handler", Category: "Security", HTTPStatus: 403},
		ErrorDefinition{Code: ErrHandlerNotFound, NumericCode: CodeHandlerNotFound, Message: handlerMessages[ErrHandlerNotFound], Layer: "Handler", Category: "Data", HTTPStatus: 404},
		ErrorDefinition{Code: ErrHandlerConflict, NumericCode: CodeHandlerConflict, Message: handlerMessages[ErrHandlerConflict], Layer: "Handler", Category: "Business", HTTPStatus: 409},
		ErrorDefinition{Code: ErrHandlerTooManyRequests, NumericCode: CodeHandlerTooManyRequests, Message: handlerMessages[ErrHandlerTooManyRequests], Layer: "Handler", Category: "RateLimit", HTTPStatus: 429},
		ErrorDefinition{Code: ErrHandlerInternal, NumericCode: CodeHandlerInternal, Message: handlerMessages[ErrHandlerInternal], Layer: "Handler", Category: "System", HTTPStatus: 500},
		ErrorDefinition{Code: ErrHandlerUnknown, NumericCode: CodeHandlerUnknown, Message: handlerMessages[ErrHandlerUnknown], Layer: "Handler", Category: "System", HTTPStatus: 500},
	)

	return defs
}

// GetErrorsByFilter filters errors by optional criteria
func GetErrorsByFilter(layer, category, code string) []ErrorDefinition {
	all := GetAllErrors()
	filtered := make([]ErrorDefinition, 0)

	for _, err := range all {
		// Filter by Layer (Type)
		if layer != "" && !strings.EqualFold(err.Layer, layer) {
			continue
		}
		// Filter by Category
		if category != "" && !strings.EqualFold(err.Category, category) {
			continue
		}
		// Filter by Code (partial match supported)
		if code != "" && !strings.Contains(strings.ToLower(err.Code), strings.ToLower(code)) {
			continue
		}

		filtered = append(filtered, err)
	}

	return filtered
}

// ... existing helper functions (getUserMessage, etc.) ...
// We need to re-implement them fully because the file is being overwritten
// I'll grab the rest of the file content from the previous `read_file` output and append it.

// Helper functions (re-implementing to ensure file completeness)

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
	// Handler layer
	case ErrHandlerBadRequest:
		return "Bad request"
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

	// Service Layer Authz
	case ErrServicePolicyViolation:
		return "Action denied by security policy"

	// Handler layer
	case ErrHandlerUnauthorized:
		return "Authentication required"
	case ErrHandlerForbidden:
		return "Access denied"
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
	// Repository layer
	case ErrRepoNotFound:
		return "Not found"
	case ErrRepoAlreadyExists:
		return "Already exists"
	// Service layer
	case ErrServiceNotFound:
		return "Not found"
	case ErrServiceAlreadyExists:
		return "Already exists"
	case ErrServiceValidation:
		return "Validation failed"
	case ErrServiceInvalidInput:
		return "Invalid input provided"
	case ErrServiceUnauthorized:
		return "Authentication required"
	case ErrServiceForbidden:
		return "Permission denied"
	case ErrServiceConflict:
		return "Resource conflict"
	// Service Layer Authz Resources
	case ErrServiceRoleNotFound:
		return "Role not found"
	case ErrServicePermissionNotFound:
		return "Permission not found"
	case ErrServiceScopeNotFound:
		return "Scope not found"

	// Handler layer
	case ErrHandlerNotFound:
		return "Not found"
	case ErrHandlerConflict:
		return "Resource conflict"
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
	// Repository layer
	case ErrRepoDatabase:
		return "Database error"
	case ErrRepoTimeout:
		return "Request timeout"
	case ErrRepoConnection:
		return "Database error"
	case ErrRepoTransaction:
		return "Database error"
	case ErrRepoConstraint:
		return "Database error"
	case ErrRepoUnknown:
		return "Unknown error"
	// Service layer
	case ErrServiceBusinessRule:
		return "Business rule violation"
	case ErrServiceDependency:
		return "Dependency error"
	case ErrServiceUnknown:
		return "Unknown error"
	// Handler layer
	case ErrHandlerInternal:
		return "Internal error"
	case ErrHandlerUnknown:
		return "Unknown error"
	default:
		return ""
	}
}

// GetNumericCode returns numeric code by error type
func GetNumericCode(code string) string {
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
	if c := GetDomainNumericCode(code); c != "" {
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
		ErrRepoConnection:    CodeRepoConnection,
		ErrRepoTimeout:       CodeRepoTimeout,
		ErrRepoTransaction:   CodeRepoTransaction,
		ErrRepoUnknown:       CodeRepoUnknown,
	}

	serviceNumericCodes = map[string]string{
		ErrServiceInvalidInput:       CodeServiceInvalidInput,
		ErrServiceValidation:         CodeServiceValidation,
		ErrServiceNotFound:           CodeServiceNotFound,
		ErrServiceAlreadyExists:      CodeServiceAlreadyExists,
		ErrServiceUnauthorized:       CodeServiceUnauthorized,
		ErrServiceForbidden:          CodeServiceForbidden,
		ErrServiceConflict:           CodeServiceConflict,
		ErrServiceBusinessRule:       CodeServiceBusinessRule,
		ErrServiceDependency:         CodeServiceDependency,
		ErrServiceUnknown:            CodeServiceUnknown,
		ErrServiceRoleNotFound:       CodeServiceRoleNotFound,
		ErrServicePermissionNotFound: CodeServicePermissionNotFound,
		ErrServicePolicyViolation:    CodeServicePolicyViolation,
		ErrServiceScopeNotFound:      CodeServiceScopeNotFound,
	}

	handlerNumericCodes = map[string]string{
		ErrHandlerBadRequest:       CodeHandlerBadRequest,
		ErrHandlerUnauthorized:     CodeHandlerUnauthorized,
		ErrHandlerForbidden:        CodeHandlerForbidden,
		ErrHandlerNotFound:         CodeHandlerNotFound,
		ErrHandlerConflict:         CodeHandlerConflict,
		ErrHandlerTooManyRequests:    CodeHandlerTooManyRequests,
		ErrHandlerNotImplemented:    CodeHandlerNotImplemented,
		ErrHandlerServiceUnavailable: CodeHandlerServiceUnavailable,
		ErrHandlerInternal:          CodeHandlerInternal,
		ErrHandlerUnknown:           CodeHandlerUnknown,
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
