package domain

import shared "gct/internal/shared/domain"

// Domain errors for the integration bounded context.
// Matched by application-layer handlers to produce appropriate HTTP status codes.
var (
	ErrIntegrationNotFound = shared.NewDomainError("INTEGRATION_NOT_FOUND", "integration not found")
)
