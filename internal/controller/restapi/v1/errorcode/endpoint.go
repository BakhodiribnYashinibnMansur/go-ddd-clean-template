package errorcode

import (
	"github.com/gin-gonic/gin"
)

func Route(api *gin.RouterGroup, c *Controller, authMiddleware, authzMiddleware, csrfMiddleware gin.HandlerFunc) {
	group := api.Group("/error-codes")
	group.Use(authMiddleware)
	group.Use(authzMiddleware)
	group.Use(csrfMiddleware)
	{
		group.POST("", c.Create)
		group.GET("", c.Gets)
		group.GET("/:code", c.Get)
		group.PUT("/:code", c.Update) // Alias for PATCH to support Swagger spec
	}
}
