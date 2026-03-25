package domain

import shared "gct/internal/shared/domain"

var (
	ErrFeatureFlagNotFound = shared.NewDomainError("FEATURE_FLAG_NOT_FOUND", "feature flag not found")
)
