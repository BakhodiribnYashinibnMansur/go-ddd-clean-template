package errors

import (
	"context"
)

// ============================================================================
// Factory Functions for Standardized Error Creation
// ============================================================================

// NewBadRequest creates a 400 Bad Request error
func NewBadRequest(ctx context.Context, message string) *AppError {
	return AutoSource(New(ctx, ErrBadRequest, message))
}

// NewUnauthorized creates a 401 Unauthorized error
func NewUnauthorized(ctx context.Context, message string) *AppError {
	if message == "" {
		message = "Authentication required"
	}
	return AutoSource(New(ctx, ErrUnauthorized, message))
}

// NewForbidden creates a 403 Forbidden error
func NewForbidden(ctx context.Context, message string) *AppError {
	if message == "" {
		message = "Access denied"
	}
	return AutoSource(New(ctx, ErrForbidden, message))
}

// NewNotFound creates a 404 Not Found error
func NewNotFound(ctx context.Context, message string) *AppError {
	if message == "" {
		message = "Resource not found"
	}
	return AutoSource(New(ctx, ErrNotFound, message))
}

// NewConflict creates a 409 Conflict error
func NewConflict(ctx context.Context, message string) *AppError {
	if message == "" {
		message = "Resource already exists"
	}
	return AutoSource(New(ctx, ErrAlreadyExists, message))
}

// NewValidationError creates a 400 Validation error
func NewValidationError(ctx context.Context, message string) *AppError {
	if message == "" {
		message = "Validation failed"
	}
	return AutoSource(New(ctx, ErrValidation, message))
}

// NewInternalError creates a 500 Internal Server error
func NewInternalError(ctx context.Context, message string) *AppError {
	if message == "" {
		message = "An unexpected error occurred"
	}
	return AutoSource(New(ctx, ErrInternal, message))
}

// NewTimeoutError creates a 504 Timeout error
func NewTimeoutError(ctx context.Context, message string) *AppError {
	if message == "" {
		message = "The operation timed out"
	}
	return AutoSource(New(ctx, ErrTimeout, message))
}

// ============================================================================
// Wrapping Functions
// ============================================================================

// WrapBadRequest wraps an error as 400 Bad Request
func WrapBadRequest(ctx context.Context, err error, message string) *AppError {
	return AutoSource(Wrap(ctx, err, ErrBadRequest, message))
}

// WrapUnauthorized wraps an error as 401 Unauthorized
func WrapUnauthorized(ctx context.Context, err error, message string) *AppError {
	return AutoSource(Wrap(ctx, err, ErrUnauthorized, message))
}

// WrapForbidden wraps an error as 403 Forbidden
func WrapForbidden(ctx context.Context, err error, message string) *AppError {
	return AutoSource(Wrap(ctx, err, ErrForbidden, message))
}

// WrapNotFound wraps an error as 404 Not Found
func WrapNotFound(ctx context.Context, err error, message string) *AppError {
	return AutoSource(Wrap(ctx, err, ErrNotFound, message))
}

// WrapConflict wraps an error as 409 Conflict
func WrapConflict(ctx context.Context, err error, message string) *AppError {
	return AutoSource(Wrap(ctx, err, ErrAlreadyExists, message))
}

// WrapInternal wraps an error as 500 Internal Server Error
func WrapInternal(ctx context.Context, err error, message string) *AppError {
	return AutoSource(Wrap(ctx, err, ErrInternal, message))
}
