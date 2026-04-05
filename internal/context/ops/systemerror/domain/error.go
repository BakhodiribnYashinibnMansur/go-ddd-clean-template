package domain

import shared "gct/internal/kernel/domain"

// Sentinel domain errors for the SystemError bounded context.
var (
	ErrSystemErrorNotFound = shared.NewDomainError("SYSTEM_ERROR_NOT_FOUND", "system error not found")
)
