package domain

import shared "gct/internal/kernel/domain"

// Domain errors for the integration bounded context.
// Matched by application-layer handlers to produce appropriate HTTP status codes.
var (
	ErrIntegrationNotFound = shared.NewDomainError("INTEGRATION_NOT_FOUND", "integration not found")
	ErrAPIKeyNotFound      = shared.NewDomainError("API_KEY_NOT_FOUND", "api key not found")
	ErrAPIKeyInactive      = shared.NewDomainError("API_KEY_INACTIVE", "api key is inactive")
)
