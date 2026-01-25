package systemerror

import (
	"gct/internal/domain"
)

// SystemError type alias or just use domain.SystemError directly
type SystemError = domain.SystemError

// CreateSystemErrorInput type alias
type CreateSystemErrorInput = domain.CreateSystemErrorInput

// ListFilter represents filter options for listing system errors
type ListFilter = domain.SystemErrorsFilter
