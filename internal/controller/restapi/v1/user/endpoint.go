package user

import (
	"gct/internal/controller/restapi/v1/user/client"
	"gct/internal/controller/restapi/v1/user/session"
	"gct/internal/controller/restapi/v1/user/setting"

	"github.com/gin-gonic/gin"
)

// UserRoute defines the routing structure for user-centric domains.
func UserRoute(api *gin.RouterGroup, controller *Controller, authMiddleware, authzMiddleware, csrfMiddleware gin.HandlerFunc) {
	client.Route(api, controller.ClientI, authMiddleware, authzMiddleware, csrfMiddleware)
	session.Route(api, controller.SessionI, authMiddleware, authzMiddleware, csrfMiddleware)
	setting.Route(api, controller.SettingI, authMiddleware)
}
