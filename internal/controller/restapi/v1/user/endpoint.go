package user

import (
	"github.com/gin-gonic/gin"

	"gct/internal/controller/restapi/v1/user/client"
	"gct/internal/controller/restapi/v1/user/session"
)

func UserRoute(api *gin.RouterGroup, controller *Controller, authMiddleware gin.HandlerFunc) {
	group := api.Group("/")

	client.Route(group, controller.ClientI, authMiddleware)
	session.Route(group, controller.SessionI, authMiddleware)
}
