package entity

import shared "gct/internal/kernel/domain"

// Domain errors for the file bounded context.
// These are returned by repositories and matched by application-layer handlers to produce appropriate HTTP responses.
var (
	ErrFileNotFound = shared.NewDomainError("FILE_NOT_FOUND", "file not found")
)
