package domain

import "fmt"

// DomainError represents a domain-level error with a code and message.
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

// Is checks if the target error is a DomainError with the same code.
func (e *DomainError) Is(target error) bool {
	t, ok := target.(*DomainError)
	if !ok {
		return false
	}
	return e.code == t.code
}
