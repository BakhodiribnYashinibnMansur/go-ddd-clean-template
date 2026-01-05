package featureflag

import (
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
)

// NewRouter creates routes for example endpoints.
func NewRouter(handler *gin.RouterGroup, l logger.Log) {
	// Feature flag examples
	ffController := NewFeatureFlagController(l.GetZap())

	ffGroup := handler.Group("")
	{
		ffGroup.GET("/boolean", ffController.ExampleBooleanFlag)
		ffGroup.GET("/string", ffController.ExampleStringVariation)
		ffGroup.GET("/int", ffController.ExampleIntVariation)
		ffGroup.GET("/json", ffController.ExampleJSONVariation)
		ffGroup.GET("/targeting", ffController.ExampleUserTargeting)
		ffGroup.GET("/rollout", ffController.ExamplePercentageRollout)
	}
}
