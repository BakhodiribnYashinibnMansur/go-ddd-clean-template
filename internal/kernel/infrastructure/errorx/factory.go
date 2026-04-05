package errorx

// ============================================================================
// Factory Functions for Standardized Error Creation
// ============================================================================

// NewBadRequest creates a 400 Bad Request error
func NewBadRequest(message string) *AppError {
	return AutoSource(New(ErrBadRequest, message))
}

// NewUnauthorized creates a 401 Unauthorized error
func NewUnauthorized(message string) *AppError {
	if message == "" {
		message = "Authentication required"
	}
	return AutoSource(New(ErrUnauthorized, message))
}

// NewForbidden creates a 403 Forbidden error
func NewForbidden(message string) *AppError {
	if message == "" {
		message = "Access denied"
	}
	return AutoSource(New(ErrForbidden, message))
}

// NewNotFound creates a 404 Not Found error
func NewNotFound(message string) *AppError {
	if message == "" {
		message = "Resource not found"
	}
	return AutoSource(New(ErrNotFound, message))
}

// NewConflict creates a 409 Conflict error
func NewConflict(message string) *AppError {
	if message == "" {
		message = "Resource already exists"
	}
	return AutoSource(New(ErrAlreadyExists, message))
}

// NewValidationError creates a 400 Validation error
func NewValidationError(message string) *AppError {
	if message == "" {
		message = "Validation failed"
	}
	return AutoSource(New(ErrValidation, message))
}

// NewInternalError creates a 500 Internal Server error
func NewInternalError(message string) *AppError {
	if message == "" {
		message = "An unexpected error occurred"
	}
	return AutoSource(New(ErrInternal, message))
}

// NewTimeoutError creates a 504 Timeout error
func NewTimeoutError(message string) *AppError {
	if message == "" {
		message = "The operation timed out"
	}
	return AutoSource(New(ErrTimeout, message))
}

// ============================================================================
// Wrapping Functions
// ============================================================================

// WrapBadRequest wraps an error as 400 Bad Request
func WrapBadRequest(err error, message string) *AppError {
	return AutoSource(Wrap(err, ErrBadRequest, message))
}

// WrapUnauthorized wraps an error as 401 Unauthorized
func WrapUnauthorized(err error, message string) *AppError {
	return AutoSource(Wrap(err, ErrUnauthorized, message))
}

// WrapForbidden wraps an error as 403 Forbidden
func WrapForbidden(err error, message string) *AppError {
	return AutoSource(Wrap(err, ErrForbidden, message))
}

// WrapNotFound wraps an error as 404 Not Found
func WrapNotFound(err error, message string) *AppError {
	return AutoSource(Wrap(err, ErrNotFound, message))
}

// WrapConflict wraps an error as 409 Conflict
func WrapConflict(err error, message string) *AppError {
	return AutoSource(Wrap(err, ErrAlreadyExists, message))
}

// WrapInternal wraps an error as 500 Internal Server Error
func WrapInternal(err error, message string) *AppError {
	return AutoSource(Wrap(err, ErrInternal, message))
}
