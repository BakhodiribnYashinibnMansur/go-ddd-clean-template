package errors

import "strings"

// ErrorDefinition represents a single error definition in the system
type ErrorDefinition struct {
	Code        string `json:"code"`
	NumericCode string `json:"numeric_code"`
	Message     string `json:"message"`
	Layer       string `json:"layer"`    // Repository, Service, Handler
	Category    string `json:"category"` // Data, Validation, Security, System, Business
	HTTPStatus  int    `json:"http_status,omitempty"`
}

// GetAllErrors returns a list of all defined errors in the system
func GetAllErrors() []ErrorDefinition {
	// Initialize with capacity for known errors
	defs := make([]ErrorDefinition, 0, 30)

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
