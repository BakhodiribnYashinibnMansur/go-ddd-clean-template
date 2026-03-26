package domain

import "fmt"

// DomainError represents a domain-level invariant violation or business rule failure.
// The code field acts as a machine-readable discriminator (e.g., "USER_NOT_FOUND") that the
// presentation layer maps to HTTP status codes. Use errors.Is for comparison — it matches on code, not message.
type DomainError struct {
	code    string
	message string
}

// NewDomainError creates a new DomainError.
func NewDomainError(code, message string) *DomainError {
	return &DomainError{
		code:    code,
		message: message,
	}
}

// Error returns the formatted error string.
func (e *DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.code, e.message)
}

// Code returns the error code.
func (e *DomainError) Code() string {
	return e.code
}

// Is checks if the target error is a DomainError with the same code, enabling errors.Is semantics.
// This allows sentinel domain errors (e.g., ErrUserNotFound) to match wrapped instances.
func (e *DomainError) Is(target error) bool {
	t, ok := target.(*DomainError)
	if !ok {
		return false
	}
	return e.code == t.code
}
