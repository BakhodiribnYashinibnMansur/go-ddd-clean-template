package translation

import (
	"github.com/gin-gonic/gin"
)

// TranslationRoute registers translation endpoints under /translations.
func TranslationRoute(api *gin.RouterGroup, ctrl ControllerI, authMiddleware, authzMiddleware gin.HandlerFunc) {
	g := api.Group("/translations", authMiddleware, authzMiddleware)
	{
		g.PUT("/:entity_type/:entity_id", ctrl.Upsert)
		g.GET("/:entity_type/:entity_id", ctrl.Gets)
		g.DELETE("/:entity_type/:entity_id", ctrl.Delete)
	}
}
