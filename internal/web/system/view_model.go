package system

import (
	apperrors "gct/pkg/errors"
)

type ErrorFilter struct {
	Layer    string `example:"Repository" form:"type"`     // Maps to 'layer'
	Category string `example:"Validation" form:"category"` // Maps to 'category'
	Code     string `example:"NOT_FOUND"  form:"code"`     // Maps to 'code'
}

type CategoryData struct {
	Name   string
	Errors []apperrors.ErrorDefinition
}

type PageData struct {
	Filter     ErrorFilter
	Categories []CategoryData
}

// SystemErrorsResponse is a wrapper for system errors list
// swagger:model SystemErrorsResponse
type SystemErrorsResponse []apperrors.ErrorDefinition
