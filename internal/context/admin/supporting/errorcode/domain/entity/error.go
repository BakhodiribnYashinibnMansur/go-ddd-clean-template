package entity

import shared "gct/internal/kernel/domain"

// Domain errors for the errorcode bounded context.
var (
	// ErrErrorCodeNotFound signals that no error code definition exists for the requested identifier.
	// Repository implementations must return this sentinel so the application layer can map it to HTTP 404.
	ErrErrorCodeNotFound = shared.NewDomainError("ERROR_CODE_NOT_FOUND", "error code not found")
)
