package domain

import shared "gct/internal/shared/domain"

var (
	ErrIntegrationNotFound = shared.NewDomainError("INTEGRATION_NOT_FOUND", "integration not found")
)
