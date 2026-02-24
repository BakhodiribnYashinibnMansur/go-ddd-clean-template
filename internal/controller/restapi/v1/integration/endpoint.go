package integration

import (
	"github.com/gin-gonic/gin"
)

// IntegrationRoute registers integration-related routes.
func IntegrationRoute(api *gin.RouterGroup, controller ControllerI, authMiddleware, authzMiddleware gin.HandlerFunc) {
	integration := api.Group("/integrations", authMiddleware, authzMiddleware)
	{
		integration.POST("", controller.CreateIntegration)
		integration.GET("", controller.ListIntegrations)
		integration.GET("/:id", controller.GetIntegration)
		integration.PUT("/:id", controller.UpdateIntegration)
		integration.DELETE("/:id", controller.DeleteIntegration)

		// API Keys sub-resources
		integration.POST("/:id/toggle", controller.ToggleIntegration)
		integration.POST("/:id/keys", controller.CreateAPIKey)
		integration.GET("/:id/keys", controller.ListAPIKeys)
	}

	keys := api.Group("/api-keys", authMiddleware)
	{
		keys.GET("/:id", controller.GetAPIKey)
		keys.POST("/:id/revoke", controller.RevokeAPIKey)
		keys.DELETE("/:id", controller.DeleteAPIKey)
	}
}
