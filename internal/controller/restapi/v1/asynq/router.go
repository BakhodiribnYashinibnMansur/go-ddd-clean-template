package asynqController

import (
	"gct/pkg/asynq"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
)

// NewRouter creates asynq routes.
func NewRouter(
	handler *gin.RouterGroup,
	asynqClient *asynq.Client,
	log logger.Log,
) {
	ctrl := NewController(asynqClient, log)

	asynqGroup := handler.Group("/asynq")
	{
		// Email endpoints
		asynqGroup.POST("/email/test", ctrl.SendTestEmail)

		// Notification endpoints
		asynqGroup.POST("/notification/test", ctrl.SendTestNotification)

		// System endpoints
		asynqGroup.POST("/seed", ctrl.SeedDatabase)
	}
}
