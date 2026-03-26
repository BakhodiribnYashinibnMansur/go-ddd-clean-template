package domain

import shared "gct/internal/shared/domain"

// Domain errors for the featureflag bounded context.
var (
	// ErrFeatureFlagNotFound signals that no feature flag exists for the requested identifier.
	// Repository implementations must return this sentinel so the application layer can map it to HTTP 404.
	ErrFeatureFlagNotFound = shared.NewDomainError("FEATURE_FLAG_NOT_FOUND", "feature flag not found")
)
