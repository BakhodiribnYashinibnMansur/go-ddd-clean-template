package featureflag

import (
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
)

// NewRouter registers the variety of feature flag demonstration endpoints.
// It maps specific flag types and evaluation strategies (Boolean, Variance, Targeting) to handlers.
func NewRouter(handler *gin.RouterGroup, l logger.Log) {
	// Initialize the controller with the underlying Zap logger instance.
	ffController := NewFeatureFlagController(l.GetZap())

	// Define the base group for feature flag experimentation.
	ffGroup := handler.Group("/featureflag")
	{
		ffGroup.GET("/boolean", ffController.ExampleBooleanFlag)       // Toggle logic test.
		ffGroup.GET("/string", ffController.ExampleStringVariation)    // A/B theme/copy test.
		ffGroup.GET("/int", ffController.ExampleIntVariation)          // Numeric tuning test.
		ffGroup.GET("/json", ffController.ExampleJSONVariation)        // Complex config test.
		ffGroup.GET("/targeting", ffController.ExampleUserTargeting)   // Identity-based evaluation.
		ffGroup.GET("/rollout", ffController.ExamplePercentageRollout) // Canary deployment simulation.
	}
}
