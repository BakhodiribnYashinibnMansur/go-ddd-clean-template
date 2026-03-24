package featureflag

import (
	"gct/internal/shared/infrastructure/logger"
)

// FeatureFlagController serves as a playground and reference implementation for the feature flag system.
type FeatureFlagController struct {
	logger logger.Log
}

// NewFeatureFlagController instantiates the controller with a structured logger.
func NewFeatureFlagController(logger logger.Log) *FeatureFlagController {
	return &FeatureFlagController{
		logger: logger,
	}
}
